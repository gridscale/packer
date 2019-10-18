package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepLinkServerSecondaryStorage struct {
	client gsclient.ServerStorageRelationOperator
	config *Config
	ui     packer.Ui
}

func (s *stepLinkServerSecondaryStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	c := s.config
	if c.SecondaryStorage {
		// Get secondary storage UUID
		secondaryStorageUUID, ok := state.Get("secondary_storage_uuid").(string)
		if !ok {
			err := errors.New("cannot convert secondary_storage_uuid to string")
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if secondaryStorageUUID == "" {
			err := errors.New("secondary_storage_uuid is empty")
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
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
		// Link server with secondary storage
		ui.Say(fmt.Sprintf("Linking the server (%s) and the secondary storage (%s)...", serverUUID, secondaryStorageUUID))
		err := client.LinkStorage(context.Background(), serverUUID, secondaryStorageUUID, false)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error linking Server with Storage: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("server_secondary_storage_linked", true)
		ui.Say(fmt.Sprintf("Linked the server (%s) and the secondary storage (%s)...", serverUUID, secondaryStorageUUID))
		return multistep.ActionContinue
	}
	ui.Say("No secondary storage is requested. Skipping linking server with secondary storage...")
	return multistep.ActionContinue
}

func (s *stepLinkServerSecondaryStorage) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	c := s.config
	if c.SecondaryStorage {
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
		// Get secondary storage UUID
		secondaryStorageUUID, ok := state.Get("secondary_storage_uuid").(string)
		if !ok {
			err := errors.New("cannot convert secondary_storage_uuid to string")
			ui.Error(err.Error())
			state.Put("error", err)
			return
		}
		if secondaryStorageUUID == "" {
			err := errors.New("secondary_storage_uuid is empty")
			ui.Error(err.Error())
			state.Put("error", err)
			return
		}
		// Unlink server and secondary storage
		serverBootStorageLinked, ok := state.Get("server_secondary_storage_linked").(bool)
		if !ok {
			err := errors.New("cannot convert server_secondary_storage_linked to boolean")
			ui.Error(err.Error())
			state.Put("error", err)
			return
		}
		if !serverBootStorageLinked {
			ui.Say("the server is not linked with the secondary storage.")
			return
		}
		ui.Say(fmt.Sprintf("Unlinking the server (%s) and the secondary storage (%s)...", serverUUID, secondaryStorageUUID))
		err := client.UnlinkStorage(context.Background(), serverUUID, secondaryStorageUUID)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error unlink server (%s) and secondary storage (%s). Please unlink them manually: %s", serverUUID, secondaryStorageUUID, err))
			return
		}
		ui.Say(fmt.Sprintf("Unlinked the server (%s) and the secondary storage (%s)", serverUUID, secondaryStorageUUID))
		return
	}
	ui.Say("No secondary storage is requested. Skipping unlinking server with secondary storage...")
}
