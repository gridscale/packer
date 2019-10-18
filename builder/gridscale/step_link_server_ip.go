package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepLinkServerIPAddr struct {
	client gsclient.ServerIPRelationOperator
	config *Config
	ui     packer.Ui
}

func (s *stepLinkServerIPAddr) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	// Get IP address UUID
	ipAddrUUID, ok := state.Get("ip_addr_uuid").(string)
	if !ok {
		err := errors.New("cannot convert ip_addr_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if ipAddrUUID == "" {
		err := errors.New("ip_addr_uuid is empty")
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
	// Link server with IP address
	ui.Say(fmt.Sprintf("Linking the server (%s) and the IP address (%s)...", serverUUID, ipAddrUUID))
	err := client.LinkIP(context.Background(), serverUUID, ipAddrUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error linking Server with IP address: %s", err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("server_ip_addr_linked", true)
	ui.Say(fmt.Sprintf("Linked the server (%s) and the IP address (%s)", serverUUID, ipAddrUUID))
	return multistep.ActionContinue
}

func (s *stepLinkServerIPAddr) Cleanup(state multistep.StateBag) {
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
	// Get IP address UUID
	ipAddrUUID, ok := state.Get("ip_addr_uuid").(string)
	if !ok {
		err := errors.New("cannot convert ip_addr_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if ipAddrUUID == "" {
		err := errors.New("ip_addr_uuid is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// Unlink server and IP address
	serverIPAddrLinked, ok := state.Get("server_ip_addr_linked").(bool)
	if !ok {
		err := errors.New("cannot convert server_ip_addr_linked to boolean")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if !serverIPAddrLinked {
		ui.Say("the server is not linked with the IP address.")
		return
	}
	ui.Say(fmt.Sprintf("Unlinking the server (%s) and the IP address (%s)...", serverUUID, ipAddrUUID))
	err := client.UnlinkIP(context.Background(), serverUUID, ipAddrUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error unlink server (%s) and IP address (%s). Please unlink them manually: %s", serverUUID, ipAddrUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Unlinked the server (%s) and the IP address (%s)", serverUUID, ipAddrUUID))
}
