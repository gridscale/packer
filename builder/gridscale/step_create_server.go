package gridscale

import (
	"context"
	"errors"
	"fmt"
	"github.com/gridscale/gsclient-go/v3"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateServer struct {
	client gsclient.ServerOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	c := s.config
	ui := s.ui
	ui.Say("Creating a new server...")
	if c.ServerName == "" {
		c.ServerName = fmt.Sprintf("packer-%s", randString(10))
	}
	log.Println(c.ServerName)
	server, err := client.CreateServer(
		context.Background(),
		gsclient.ServerCreateRequest{
			Name:   c.ServerName,
			Cores:  c.ServerCores,
			Memory: c.ServerMemory,
		})
	if err != nil {
		err := fmt.Errorf("Error creating server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("server_uuid", server.ObjectUUID)
	ui.Say(fmt.Sprintf("a server (%s) has been created", server.ObjectUUID))
	return multistep.ActionContinue
}

func (s *stepCreateServer) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	serverUUID, ok := state.Get("server_uuid").(string)
	if !ok {
		err := errors.New("cannot convert server_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if serverUUID == "" {
		ui.Say("No server UUID detected.")
		return
	}
	// Shutdown the server before removing it
	ui.Say(fmt.Sprintf("Shutting down the server (%s)...", serverUUID))
	err := client.StopServer(context.Background(), serverUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error shutdown server: %s", err))
		return
	}
	ui.Say(fmt.Sprintf("Shutted down the server (%s)", serverUUID))
	// Remove the server
	ui.Say(fmt.Sprintf("Destroying the server (%s)...", serverUUID))
	err = client.DeleteServer(context.Background(), serverUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying server. Please destroy it manually: %s", err))
		return
	}
	ui.Say(fmt.Sprintf("Destroyed the server (%s)", serverUUID))
}
