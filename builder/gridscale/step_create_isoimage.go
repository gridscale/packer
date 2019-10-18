package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateISOImage struct {
	client gsclient.ISOImageOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateISOImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	c := s.config
	ui := s.ui
	// If IsoImageUUID is set, use it instead of creating a new ISO-image
	if c.IsoImageUUID != "" {
		ui.Say("Getting ISO-image UUID from config...")
		state.Put("iso_image_uuid", c.IsoImageUUID)
		return multistep.ActionContinue
	}
	if c.IsoImageURL != "" {
		ui.Say("Creating an ISO-image...")
		isoImageCreateRequest := gsclient.ISOImageCreateRequest{
			Name:      c.ServerName,
			SourceURL: c.IsoImageURL,
		}
		isoImage, err := client.CreateISOImage(context.Background(), isoImageCreateRequest)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error creating ISO-image: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("iso_image_uuid", isoImage.ObjectUUID)
		ui.Say(fmt.Sprintf("an ISO-image (%s) has been created", isoImage.ObjectUUID))
		return multistep.ActionContinue
	}
	ui.Say("No ISO-image is requested. Skipping creating an ISO-image...")
	return multistep.ActionContinue
}

func (s *stepCreateISOImage) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	c := s.config
	// If IsoImageUUID is set, skip removing ISO-image
	if c.IsoImageUUID != "" {
		return
	}
	// If ISO-image is not used, skip.
	if c.IsoImageUUID == "" && c.IsoImageURL == "" {
		ui.Say("No ISO-image is requested. Skipping removing ISO-image...")
		return
	}
	// Destroy the requested ISO-image
	isoImageUUID, ok := state.Get("iso_image_uuid").(string)
	if !ok {
		err := errors.New("cannot convert iso_image_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if isoImageUUID == "" {
		ui.Say("No ISO-image UUID detected.")
		return
	}
	ui.Say(fmt.Sprintf("Destroying the ISO-image (%s)...", isoImageUUID))
	err := client.DeleteISOImage(context.Background(), isoImageUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying ISO-image (%s). Please destroy it manually: %s", isoImageUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Destroyed the ISO-image (%s)", isoImageUUID))
}
