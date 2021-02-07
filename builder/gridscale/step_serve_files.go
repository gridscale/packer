package gridscale

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/gridscale/easyssh"
)

const (
	ubuntuTemplateName            = "Ubuntu 20.04 LTS (Focal Fossa)"
	fileServerPlainPassword       = "file-server-gridscale"
	fileServerAddressPlaceholder  = "__HTTP__ADDRESS__"
	initialBootTimeFileServerSecs = 30
)

type fileHTTPServerCreator interface {
	CreateServer(ctx context.Context, body gsclient.ServerCreateRequest) (gsclient.ServerCreateResponse, error)
	CreateStorage(ctx context.Context, body gsclient.StorageCreateRequest) (gsclient.CreateResponse, error)
	CreateIP(ctx context.Context, body gsclient.IPCreateRequest) (gsclient.IPCreateResponse, error)
	GetTemplateByName(ctx context.Context, name string) (gsclient.Template, error)
	LinkStorage(ctx context.Context, serverID string, storageID string, bootdevice bool) error
	GetNetworkPublic(ctx context.Context) (gsclient.Network, error)
	LinkIP(ctx context.Context, serverID string, ipID string) error
	LinkNetwork(ctx context.Context, serverID, networkID, firewallTemplate string, bootdevice bool, order int, l3security []string, firewall *gsclient.FirewallRules) error
	StartServer(ctx context.Context, id string) error
	StopServer(ctx context.Context, id string) error
	DeleteServer(ctx context.Context, id string) error
	DeleteStorage(ctx context.Context, id string) error
	DeleteIP(ctx context.Context, id string) error
}

type stepServeHTTPFiles struct {
	client fileHTTPServerCreator
	config *Config
	ui     packer.Ui
}

func (s *stepServeHTTPFiles) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := s.config
	ui := s.ui
	// If a list of file is set, serve all files in the list
	if len(c.Files) > 0 {
		client := s.client
		ui.Say("Creating a HTTP server to serve files...")
		// Create a server
		serverRes, err := client.CreateServer(
			context.Background(),
			gsclient.ServerCreateRequest{
				Name:   "file-server",
				Cores:  1,
				Memory: 2,
			},
		)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error creating a HTTP server: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("file_server_uuid", serverRes.ObjectUUID)
		// Get ubuntu 20.04 template
		template, err := client.GetTemplateByName(context.Background(), ubuntuTemplateName)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error getting ubuntu template: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		// Create a bootable storage from ubuntu template
		storageRes, err := client.CreateStorage(
			context.Background(),
			gsclient.StorageCreateRequest{
				Capacity:    10,
				Name:        "file-server-storage",
				StorageType: gsclient.InsaneStorageType,
				Template: &gsclient.StorageTemplate{
					TemplateUUID: template.Properties.ObjectUUID,
					Password:     fileServerPlainPassword,
					PasswordType: gsclient.PlainPasswordType,
					Hostname:     "file-server-storage",
				},
			},
		)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error creating file server's boot storage: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("file_server_storage_uuid", storageRes.ObjectUUID)
		// Create an IPv4 address
		ipAddrRes, err := client.CreateIP(
			context.Background(),
			gsclient.IPCreateRequest{
				Name:   "file-server-IPv4",
				Family: gsclient.IPv4Type,
			},
		)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error creating file server's IPv4 address: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("file_server_ip_uuid", ipAddrRes.ObjectUUID)
		// Get public network UUID
		pubNetUUID, ok := state.Get("public_network_uuid").(string)
		if !ok {
			err := errors.New("cannot convert public_network_uuid to string")
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if pubNetUUID == "" {
			err := errors.New("public_network_uuid is empty")
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		// Link file server - boot storage
		err = client.LinkStorage(context.Background(), serverRes.ObjectUUID, storageRes.ObjectUUID, true)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error linking file server - boot storage: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		// Link file server - IPv4 address
		err = client.LinkIP(context.Background(), serverRes.ObjectUUID, ipAddrRes.ObjectUUID)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error linking file server - IPv4 address: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		// Link file server - Public network
		err = client.LinkNetwork(context.Background(), serverRes.ObjectUUID, pubNetUUID, "", false, 0, nil, nil)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error linking file Server - public network: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		// Start the file server
		err = client.StartServer(context.Background(), serverRes.ObjectUUID)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error starting file server: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		time.Sleep(time.Second * initialBootTimeFileServerSecs)
		// Create MakeConfig instance with remote username, server address..
		sshCfg := &easyssh.MakeConfig{
			User:     "root",
			Server:   ipAddrRes.IP,
			Password: fileServerPlainPassword,
			Port:     "22",
		}
		// Upload files
		err = uploadFilesToServer(sshCfg, ui, c.Files)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error uploading files to file server: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		// SSH to the file server to start serving files
		_, stderr, _, err := sshCfg.Run("nohup python3 -u -m http.server 8080 </dev/null >/dev/null 2>&1 &", 60)
		// Handle errors
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error running remote command in file server: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if strings.ReplaceAll(stderr, "\n", "") != "" {
			ui.Error(fmt.Sprintf(
				"Error running remote command in file server (stderr): %s", errors.New(stderr)))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		// PopulateProvisionHookData in github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps/step_provision.go
		// will look for `http_port` and `http_ip` to replace the placeholders {{ .HTTPIP }} and {{ .HTTPPort }} in
		// shell provisioner's command
		state.Put("http_ip", ipAddrRes.IP)
		state.Put("http_port", 8080)

		replaceFileServerPlaceholder(c, ipAddrRes.IP+":8080")

		ui.Say(fmt.Sprintf("a file server is ready at address: %s:8080", ipAddrRes.IP))
		return multistep.ActionContinue
	}
	ui.Say("No file server is requested. Skipping creating a file server...")
	return multistep.ActionContinue
}

func (s *stepServeHTTPFiles) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	c := s.config
	if len(c.Files) > 0 {
		removeFileServerResources(client, state, ui)
	}
}

func uploadFilesToServer(sshCfg *easyssh.MakeConfig, ui packer.Ui, files []string) error {
	for _, relPath := range files {
		// check if the file is regular or dir
		fileInf, err := os.Stat(relPath)
		if err != nil {
			return err
		}

		// if file is dir, get all files in that dir and upload them
		if fileInf.IsDir() {
			err := filepath.Walk(relPath, func(path string, info os.FileInfo, err error) error {
				if info.Mode().IsRegular() {
					ui.Say(fmt.Sprintf("Uploading file \"%s\"...", path))
					err = sshCfg.Scp(path, path)
					if err != nil {
						ui.Say(fmt.Sprintf("Failed to uploaded file \"%s\"!", path))
						return err
					}
					ui.Say(fmt.Sprintf("Uploaded file \"%s\" succesfully", path))
				}
				return nil
			})
			if err != nil {
				return err
			}
			continue
		}

		// if file is regular file, upload the file
		ui.Say(fmt.Sprintf("Uploading file \"%s\"...", relPath))
		err = sshCfg.Scp(relPath, relPath)
		if err != nil {
			ui.Say(fmt.Sprintf("Failed to uploaded file \"%s\"!", relPath))
			return err
		}
		ui.Say(fmt.Sprintf("Uploaded file \"%s\" succesfully", relPath))
	}
	return nil
}

func removeFileServerResources(client fileHTTPServerCreator, state multistep.StateBag, ui packer.Ui) {
	ui.Say("Destroying all resources of the file server...")
	if fileServerUUID, _ := state.Get("file_server_uuid").(string); fileServerUUID != "" {
		if err := client.StopServer(context.Background(), fileServerUUID); err != nil {
			ui.Error(fmt.Sprintf(
				"Error stopping file server: %s, please go to gridscale panel to remove this server", err))
		}
		if err := client.DeleteServer(context.Background(), fileServerUUID); err != nil {
			ui.Error(fmt.Sprintf(
				"Error removing file server: %s, please go to gridscale panel to remove it", err))
		}
		ui.Say(fmt.Sprintf("Destroyed the file server (%s)", fileServerUUID))
	}
	if fileServerStorageUUID, _ := state.Get("file_server_storage_uuid").(string); fileServerStorageUUID != "" {
		if err := client.DeleteStorage(context.Background(), fileServerStorageUUID); err != nil {
			ui.Error(fmt.Sprintf(
				"Error removing file server's storage: %s, please go to gridscale panel to remove it", err))
		}
		ui.Say(fmt.Sprintf("Destroyed the file server's storage (%s)", fileServerStorageUUID))
	}
	if fileServerIPAddrUUID, _ := state.Get("file_server_ip_uuid").(string); fileServerIPAddrUUID != "" {
		if err := client.DeleteIP(context.Background(), fileServerIPAddrUUID); err != nil {
			ui.Error(fmt.Sprintf(
				"Error removing file server's IP address: %s, please go to gridscale panel to remove it", err))
		}
		ui.Say(fmt.Sprintf("Destroyed the file server's IP address (%s)", fileServerIPAddrUUID))
	}
}

func replaceFileServerPlaceholder(cfg *Config, serverAddr string) {
	newBootCommands := []string{}
	for _, bootCmd := range cfg.BootCommand {
		newBootCommands = append(newBootCommands, strings.ReplaceAll(bootCmd, fileServerAddressPlaceholder, serverAddr))
	}
	cfg.BootCommand = newBootCommands
}
