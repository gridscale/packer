package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepLinkServerBootStorage struct {
	client gsclient.ServerStorageRelationOperator
	config *Config
	ui     packer.Ui
}

func (s *stepLinkServerBootStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	// Get boot storage UUID
	bootStorageUUID, ok := state.Get("boot_storage_uuid").(string)
	if !ok {
		err := errors.New("cannot convert boot_storage_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if bootStorageUUID == "" {
		err := errors.New("boot_storage_uuid is empty")
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
	// Link server with boot storage
	ui.Say(fmt.Sprintf("Linking the server (%s) and the boot storage (%s)...", serverUUID, bootStorageUUID))
	err := client.LinkStorage(context.Background(), serverUUID, bootStorageUUID, true)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error linking Server with Storage: %s", err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("server_boot_storage_linked", true)
	ui.Say(fmt.Sprintf("Linked the server (%s) and the boot storage (%s)", serverUUID, bootStorageUUID))
	return multistep.ActionContinue
}

func (s *stepLinkServerBootStorage) Cleanup(state multistep.StateBag) {
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
		err := errors.New("serverUUID is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// Get boot storage UUID
	bootStorageUUID, ok := state.Get("boot_storage_uuid").(string)
	if !ok {
		err := errors.New("cannot convert boot_storage_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if bootStorageUUID == "" {
		err := errors.New("boot_storage_uuid is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// Unlink server and boot storage
	serverBootStorageLinked, ok := state.Get("server_boot_storage_linked").(bool)
	if !ok {
		err := errors.New("cannot convert server_boot_storage_linked to boolean")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if !serverBootStorageLinked {
		ui.Say("the server is not linked with the boot storage.")
		return
	}
	ui.Say(fmt.Sprintf("Unlinking the server (%s) and the boot storage (%s)...", serverUUID, bootStorageUUID))
	err := client.UnlinkStorage(context.Background(), serverUUID, bootStorageUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error unlink server (%s) and boot storage (%s). Please unlink them manually: %s", serverUUID, bootStorageUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Unlinked the server (%s) and the boot storage (%s)", serverUUID, bootStorageUUID))
}
