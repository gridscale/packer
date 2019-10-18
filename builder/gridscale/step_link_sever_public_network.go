package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepLinkServerPublicNetwork struct {
	client gsclient.ServerNetworkRelationOperator
	ui     packer.Ui
}

func (s *stepLinkServerPublicNetwork) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
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
	// Link server with public network
	ui.Say(fmt.Sprintf("Linking the server (%s) and the public network (%s)...", serverUUID, pubNetUUID))
	err := client.LinkNetwork(context.Background(), serverUUID, pubNetUUID, "", false, 0, nil, nil)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error linking Server with public network: %s", err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("server_public_network_linked", true)
	ui.Say(fmt.Sprintf("Linked the server (%s) and the public network (%s)", serverUUID, pubNetUUID))
	return multistep.ActionContinue
}

func (s *stepLinkServerPublicNetwork) Cleanup(state multistep.StateBag) {
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
	// Get public network UUID
	pubNetUUID, ok := state.Get("public_network_uuid").(string)
	if !ok {
		err := errors.New("cannot convert public_network_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if pubNetUUID == "" {
		err := errors.New("public_network_uuid is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// Unlink server and public network
	serverPublicNetworkLinked, ok := state.Get("server_public_network_linked").(bool)
	if !ok {
		err := errors.New("cannot convert server_public_network_linked to boolean")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if !serverPublicNetworkLinked {
		ui.Say("the server is not linked with the public network.")
		return
	}
	ui.Say(fmt.Sprintf("Unlinking the server (%s) and the public network (%s)...", serverUUID, pubNetUUID))
	err := client.UnlinkNetwork(context.Background(), serverUUID, pubNetUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error unlink server (%s) and public network (%s). Please unlink them manually: %s", serverUUID, pubNetUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Unlinked the server (%s) and the public network (%s)", serverUUID, pubNetUUID))
}
