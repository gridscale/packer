package gridscale

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mitchellh/go-vnc"
	"golang.org/x/net/websocket"
)

type StepVNCConnect struct {
	client gsclient.ServerOperator
	config *Config
	ui     packer.Ui
}

func (s StepVNCConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	c := s.config
	if len(c.BootCommand) == 0 {
		ui.Say("boot_command is not set. Skipping connecting to VNC server...")
		return multistep.ActionContinue
	}
	ui.Say("Connecting to VNC server...")
	// Get server UUID
	serverUUID, ok := state.Get("server_uuid").(string)
	if !ok {
		err := errors.New("cannot convert server_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if serverUUID == "" {
		err := errors.New("serverUUID is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	// Get server's VNC console token
	server, err := client.GetServer(context.Background(), serverUUID)
	if err != nil {
		err := fmt.Errorf("Error getting server's VNC console token: %s\n", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	// Config the VNC console URL endpoint
	websocketUrl := fmt.Sprintf("wss://api.gridscale.io/console/?token=%s", server.Properties.ConsoleToken)
	log.Printf("[DEBUG] websocket url: %s", websocketUrl)
	u, err := url.Parse(websocketUrl)
	if err != nil {
		err := fmt.Errorf("Error parsing websocket url: %s\n", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	// Maybe for CORS
	origin, err := url.Parse("http://localhost")
	if err != nil {
		err := fmt.Errorf("Error parsing websocket origin url: %s\n", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Create the websocket connection and set it to a BinaryFrame
	websocketConfig := &websocket.Config{
		Location: u,
		Origin:   origin,
		// Not sure about TLS things
		TlsConfig: &tls.Config{InsecureSkipVerify: false},
		Version:   websocket.ProtocolVersionHybi13,
		Protocol:  []string{"binary"},
	}
	nc, err := websocket.DialConfig(websocketConfig)
	if err != nil {
		err := fmt.Errorf("Error Dialing: %s\n", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	nc.PayloadType = websocket.BinaryFrame

	// Setup the VNC connection over the websocket
	ccconfig := &vnc.ClientConfig{
		Auth:      []vnc.ClientAuth{new(vnc.ClientAuthNone)},
		Exclusive: false,
	}
	vncClient, err := vnc.Client(nc, ccconfig)
	if err != nil {
		err := fmt.Errorf("Error setting the VNC over websocket client: %s\n", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("vnc_conn", vncClient)
	ui.Say("VNC connected")
	return multistep.ActionContinue
}

func (s StepVNCConnect) Cleanup(state multistep.StateBag) {
}
