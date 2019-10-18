package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateSnapshot struct {
	client gsclient.StorageSnapshotOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	c := s.config
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

	ui.Say(fmt.Sprintf("Creating snapshot: %v", c.TemplateName))
	snapshot, err := client.CreateStorageSnapshot(
		context.Background(),
		bootStorageUUID,
		gsclient.StorageSnapshotCreateRequest{
			Name: c.TemplateName,
		})

	if err != nil {
		err := fmt.Errorf("Error creating snapshot: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("snapshot_uuid", snapshot.ObjectUUID)
	ui.Say(fmt.Sprintf("Created snapshot %v with uuid: %v", c.TemplateName, snapshot.ObjectUUID))

	return multistep.ActionContinue
}

func (s *stepCreateSnapshot) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
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
	// Get snapshot UUID
	snapshotUUID, ok := state.Get("snapshot_uuid").(string)
	if !ok {
		err := errors.New("cannot convert snapshot_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if snapshotUUID == "" {
		err := errors.New("snapshot_uuid is empty")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	// remove snapshot
	ui.Say(fmt.Sprintf("Destroying the snapshot (%s) of storage (%s)...", snapshotUUID, bootStorageUUID))
	err := client.DeleteStorageSnapshot(context.Background(), bootStorageUUID, snapshotUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying snapshot. Please destroy it manually: %s", err))
		return
	}
	ui.Say(fmt.Sprintf("Destroyed the snapshot (%s) of storage (%s)", snapshotUUID, bootStorageUUID))
}
