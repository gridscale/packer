package gridscale

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gridscale/gsclient-go/v3"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

const serverShutdownTimeoutSecs = 120

type stepStartServer struct {
	client gsclient.ServerOperator
	ui     packer.Ui
}

func (s *stepStartServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
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
	ui.Say("Starting server...")
	ui.Say(fmt.Sprintf("Starting server (%s)...", serverUUID))
	err := client.StartServer(context.Background(), serverUUID)
	if err != nil {
		err := fmt.Errorf("Error starting server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Started server (%s)", serverUUID))
	return multistep.ActionContinue
}

func (s *stepStartServer) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	// Get server UUID
	serverUUID, ok := state.Get("server_uuid").(string)
	if !ok {
		err := errors.New("cannot convert server_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if serverUUID == "" {
		err := errors.New("server_uuid is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	ui.Say(fmt.Sprintf("Gracefully shutting down server (%s)...", serverUUID))
	//set the shutdown timeout specifically
	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeoutSecs*time.Second)
	defer cancel()
	err := client.ShutdownServer(shutdownCtx, serverUUID)
	if err != nil && err != shutdownCtx.Err() {
		err := fmt.Errorf("Error shutting down server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return
	}
	// if the server cannot be shutdown gracefully, try to turn it off
	if err != nil && err == shutdownCtx.Err() {
		ui.Say(fmt.Sprintf("Could not gracefully shutdown server (%s). Trying to turn it off instead...", serverUUID))
		err := client.StopServer(context.Background(), serverUUID)
		if err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf(
				"Error shutdown server: %s", err))
			return
		}
		ui.Say(fmt.Sprintf("Successfully turned off server (%s).", serverUUID))
		return
	}
	ui.Say(fmt.Sprintf("Gracefully shut down server (%s)", serverUUID))
}
