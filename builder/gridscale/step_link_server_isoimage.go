package gridscale

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepLinkServerISOImage struct {
	client gsclient.ServerIsoImageRelationOperator
	config *Config
	ui     packer.Ui
}

func (s *stepLinkServerISOImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	c := s.config
	if c.IsoImageUUID == "" && c.IsoImageURL == "" {
		ui.Say("No ISO-image is requested. Skipping linking the server with an ISO-image...")
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
		ui.Say("No ISO-image UUID detected. Skipping linking the server with an ISO-image...")
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
	// Link server with ISO-image
	ui.Say(fmt.Sprintf("Linking the server (%s) and the ISO-image (%s)...", serverUUID, isoImageUUID))
	err := client.LinkIsoImage(context.Background(), serverUUID, isoImageUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error linking Server with Storage: %s", err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("server_iso_image_linked", true)
	ui.Say(fmt.Sprintf("Linked the server (%s) and the ISO-image (%s)", serverUUID, isoImageUUID))
	return multistep.ActionContinue
}

func (s *stepLinkServerISOImage) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	c := s.config
	if c.IsoImageUUID == "" && c.IsoImageURL == "" {
		ui.Say("No ISO-image is requested. Skipping unlinking server with an ISO-image...")
		return
	}
	// Get ISO-image UUID
	isoImageUUID, ok := state.Get("iso_image_uuid").(string)
	if !ok {
		err := errors.New("cannot convert iso_image_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if isoImageUUID == "" {
		ui.Say("No ISO-image UUID detected. Skipping unlinking the server with an ISO-image...")
		return
	}
	// Get server UUID
	serverUUID, ok := state.Get("server_uuid").(string)
	if !ok {
		err := errors.New("cannot convert server_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if serverUUID == "" {
		err := errors.New("serverUUID is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// Unlink server and ISO-image
	serverISOImageLinked, ok := state.Get("server_iso_image_linked").(bool)
	if !ok {
		err := errors.New("cannot convert server_iso_image_linked to boolean")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if !serverISOImageLinked {
		ui.Say("the server is not linked with the ISO-image.")
		return
	}
	ui.Say(fmt.Sprintf("Unlinking the server (%s) and the ISO-image (%s)...", serverUUID, isoImageUUID))
	err := removeErrorContainsHTTPCodes(
		client.UnlinkIsoImage(context.Background(), serverUUID, isoImageUUID),
		http.StatusConflict,
		http.StatusNotFound,
	)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error unlink server (%s) and ISO-image (%s). Please unlink them manually: %s", serverUUID, isoImageUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Unlinked the server (%s) and the ISO-image (%s)", serverUUID, isoImageUUID))
}
