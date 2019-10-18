package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateTemplate struct {
	client gsclient.TemplateOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	c := s.config
	ui := s.ui

	// Get snapshot UUID
	snapshotUUID, ok := state.Get("snapshot_uuid").(string)
	if !ok {
		err := errors.New("cannot convert snapshot_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if snapshotUUID == "" {
		err := errors.New("snapshot_uuid is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Creating template: %v", c.TemplateName))
	template, err := client.CreateTemplate(
		context.Background(),
		gsclient.TemplateCreateRequest{
			Name:         c.TemplateName,
			SnapshotUUID: snapshotUUID,
		})

	if err != nil {
		err := fmt.Errorf("Error creating template: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("template_uuid", template.ObjectUUID)
	ui.Say(fmt.Sprintf("Created template %v with uuid: %v", c.TemplateName, template.ObjectUUID))
	return multistep.ActionContinue
}

func (s *stepCreateTemplate) Cleanup(state multistep.StateBag) {
	// no cleanup
}
