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

const serverShutdownPostInstallationTimeoutSecs = 30

type stepCleanupISOImageInstallation struct {
	relClient gsclient.ServerIsoImageRelationOperator
	sClient   gsclient.ServerOperator
	config    *Config
	ui        packer.Ui
}

func (s stepCleanupISOImageInstallation) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	relClient := s.relClient
	ui := s.ui
	c := s.config
	if c.IsoImageUUID == "" && c.IsoImageURL == "" {
		ui.Say("No ISO-image is requested. Skipping cleaning ISO-image installation...")
		return multistep.ActionContinue
	}
	// Get ISO-image UUID
	isoImageUUID, ok := state.Get("iso_image_uuid").(string)
	if !ok {
		err := errors.New("cannot convert iso_image_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if isoImageUUID == "" {
		ui.Say("No ISO-image UUID detected. Skipping cleaning ISO-image installation...")
		return multistep.ActionContinue
	}
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

	// Unlink server and ISO-image
	serverISOImageLinked, ok := state.Get("server_iso_image_linked").(bool)
	if !ok {
		err := errors.New("cannot convert server_iso_image_linked to boolean")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if !serverISOImageLinked {
		ui.Say("the server is not linked with the ISO-image.")
		return multistep.ActionContinue
	}
	ui.Say(fmt.Sprintf("Cleaning up the ISO-image (%s) installation on server (%s)...", isoImageUUID, serverUUID))
	err := relClient.UnlinkIsoImage(context.Background(), serverUUID, isoImageUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error unlink server (%s) and ISO-image (%s). Please unlink them manually: %s", serverUUID, isoImageUUID, err))
		return multistep.ActionHalt
	}

	// Restart the server
	sClient := s.sClient
	// set the shutdown timeout specifically
	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownPostInstallationTimeoutSecs*time.Second)
	defer cancel()
	err = sClient.ShutdownServer(shutdownCtx, serverUUID)
	if err != nil && err != shutdownCtx.Err() {
		err := fmt.Errorf("Error shutting down server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// if the server cannot be shutdown gracefully, try to turn it off
	if err != nil && err == shutdownCtx.Err() {
		err := sClient.StopServer(context.Background(), serverUUID)
		if err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf(
				"Error shutdown server: %s", err))
			return multistep.ActionHalt
		}
	}
	// Start the server
	err = sClient.StartServer(context.Background(), serverUUID)
	if err != nil {
		err := fmt.Errorf("Error starting server: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Cleaned up the ISO-image (%s) installation on server (%s)", isoImageUUID, serverUUID))
	return multistep.ActionContinue
}

func (s stepCleanupISOImageInstallation) Cleanup(state multistep.StateBag) {
}
