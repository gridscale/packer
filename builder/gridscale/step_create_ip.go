package gridscale

import (
	"context"
	"errors"
	"fmt"
	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateIPAddr struct {
	client gsclient.IPOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateIPAddr) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	c := s.config
	ui := s.ui
	// Create an IP address
	ui.Say("Creating an IP address...")
	ip, err := client.CreateIP(
		context.Background(),
		gsclient.IPCreateRequest{
			Name:   c.ServerName,
			Family: gsclient.IPv4Type,
		})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error creating IP address: %s", err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("ip_addr_uuid", ip.ObjectUUID)
	state.Put("server_ip", ip.IP)
	ui.Say(fmt.Sprintf("an IP address %s (%s) has been created", ip.IP, ip.ObjectUUID))
	return multistep.ActionContinue
}

func (s *stepCreateIPAddr) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	// Destroy the created IP address
	ipAddrUUID, ok := state.Get("ip_addr_uuid").(string)
	if !ok {
		err := errors.New("cannot convert ip_addr_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if ipAddrUUID == "" {
		ui.Say("No IP address UUID detected.")
		return
	}
	ui.Say(fmt.Sprintf("Destroying the IP address (%s)...", ipAddrUUID))
	err := client.DeleteIP(context.Background(), ipAddrUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying IP address (%s). Please destroy it manually: %s", ipAddrUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Destroyed the IP address (%s)", ipAddrUUID))
}
