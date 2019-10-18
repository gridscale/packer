package gridscale

import (
	"context"
	"errors"
	"fmt"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepCreateSecondaryStorage struct {
	client gsclient.StorageOperator
	config *Config
	ui     packer.Ui
}

func (s *stepCreateSecondaryStorage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := s.config
	ui := s.ui
	// Create a secondary storage
	if c.SecondaryStorage == true {
		client := s.client
		ui.Say("Creating a secondary storage...")
		storage, err := client.CreateStorage(
			context.Background(),
			gsclient.StorageCreateRequest{
				Capacity:    c.StorageCapacity,
				Name:        fmt.Sprintf("%s-secondary", c.ServerName),
				StorageType: gsclient.InsaneStorageType,
			})

		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error creating secondary storage: %s", err))
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("secondary_storage_uuid", storage.ObjectUUID)
		ui.Say(fmt.Sprintf("a secondary storage (%s) has been created", storage.ObjectUUID))
		return multistep.ActionContinue
	}
	ui.Say("No secondary storage is requested. Skipping creating a secondary storage...")
	return multistep.ActionContinue
}

func (s *stepCreateSecondaryStorage) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	c := s.config
	if c.SecondaryStorage == true {
		// Destroy the created secondary storage
		secondaryStorageUUID, ok := state.Get("secondary_storage_uuid").(string)
		if !ok {
			err := errors.New("cannot convert secondary_storage_uuid to string")
			ui.Error(err.Error())
			state.Put("error", err)
			return
		}
		if secondaryStorageUUID == "" {
			ui.Say("No secondary storage UUID detected.")
			return
		}
		ui.Say(fmt.Sprintf("Destroying the secondary storage (%s)...", secondaryStorageUUID))
		err := client.DeleteStorage(context.Background(), secondaryStorageUUID)
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error destroying secondary storage (%s). Please destroy it manually: %s", secondaryStorageUUID, err))
			return
		}
		ui.Say(fmt.Sprintf("Destroyed the secondary storage (%s)", secondaryStorageUUID))
	}
}
