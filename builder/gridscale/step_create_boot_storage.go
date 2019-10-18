package gridscale

import (
	"context"
	"errors"
	"fmt"
	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateBootStorage struct {
	client gsclient.StorageOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateBootStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	c := s.config

	// Create a boot storage
	ui.Say("Creating a boot storage...")
	if c.Hostname == "" {
		c.Hostname = "packer-hostname"
	}
	storageCreateReq := gsclient.StorageCreateRequest{
		Capacity:    c.StorageCapacity,
		Name:        c.ServerName,
		StorageType: gsclient.InsaneStorageType,
	}
	if c.BaseTemplateUUID != "" {
		sshKeyUUID, ok := state.Get("ssh_key_uuid").(string)
		if !ok {
			err := errors.New("cannot convert ssh_key_uuid to string")
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if sshKeyUUID == "" {
			ui.Error("No SSH key UUID detected.")
			state.Put("error", "No SSH key UUID detected.")
			return multistep.ActionHalt
		}
		storageCreateReq.Template = &gsclient.StorageTemplate{
			Password:     c.Comm.SSHPassword,
			PasswordType: gsclient.PlainPasswordType,
			Hostname:     c.Hostname,
			Sshkeys:      []string{sshKeyUUID},
			TemplateUUID: c.BaseTemplateUUID,
		}
	}
	storage, err := client.CreateStorage(context.Background(), storageCreateReq)

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error creating boot storage: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("boot_storage_uuid", storage.ObjectUUID)
	ui.Say(fmt.Sprintf("a boot storage (%s) has been created", storage.ObjectUUID))
	return multistep.ActionContinue
}

func (s *stepCreateBootStorage) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	// Destroy the created boot storage
	bootStorageUUID, ok := state.Get("boot_storage_uuid").(string)
	if !ok {
		err := errors.New("cannot convert boot_storage_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if bootStorageUUID == "" {
		ui.Say("No boot storage UUID detected.")
		return
	}
	ui.Say(fmt.Sprintf("Destroying the boot storage (%s)...", bootStorageUUID))
	err := client.DeleteStorage(context.Background(), bootStorageUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying boot storage (%s). Please destroy it manually: %s", bootStorageUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Destroyed the boot storage (%s)", bootStorageUUID))
}
