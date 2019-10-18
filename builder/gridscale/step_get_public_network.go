package gridscale

import (
	"context"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepGetPublicNetwork struct {
	client gsclient.NetworkOperator
	ui     packer.Ui
}

func (s *stepGetPublicNetwork) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	// Get public network
	publicNetwork, err := client.GetNetworkPublic(context.Background())
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error getting public network: %s", err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("public_network_uuid", publicNetwork.Properties.ObjectUUID)
	return multistep.ActionContinue
}

func (s *stepGetPublicNetwork) Cleanup(state multistep.StateBag) {
	// no cleanup
}
