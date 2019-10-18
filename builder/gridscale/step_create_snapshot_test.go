package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type SnapshotOperatorMock struct{}

func (s SnapshotOperatorMock) GetStorageSnapshotList(ctx context.Context, id string) ([]gsclient.StorageSnapshot, error) {
	panic("implement me")
}

func (s SnapshotOperatorMock) GetSnapshotsByLocation(ctx context.Context, id string) ([]gsclient.StorageSnapshot, error) {
	panic("implement me")
}

func (s SnapshotOperatorMock) GetStorageSnapshot(ctx context.Context, storageID, snapshotID string) (gsclient.StorageSnapshot, error) {
	panic("implement me")
}

func (s SnapshotOperatorMock) CreateStorageSnapshot(ctx context.Context, id string, body gsclient.StorageSnapshotCreateRequest) (gsclient.StorageSnapshotCreateResponse, error) {
	if id == "success" {
		return gsclient.StorageSnapshotCreateResponse{
			RequestUUID: "test",
			ObjectUUID:  "test",
		}, nil
	}
	return gsclient.StorageSnapshotCreateResponse{}, errors.New("error")
}

func (s SnapshotOperatorMock) UpdateStorageSnapshot(ctx context.Context, storageID, snapshotID string, body gsclient.StorageSnapshotUpdateRequest) error {
	panic("implement me")
}

func (s SnapshotOperatorMock) DeleteStorageSnapshot(ctx context.Context, storageID, snapshotID string) error {
	if snapshotID == "success" {
		return nil
	}
	return errors.New("error")
}

func (s SnapshotOperatorMock) GetDeletedSnapshots(ctx context.Context) ([]gsclient.StorageSnapshot, error) {
	panic("implement me")
}

func (s SnapshotOperatorMock) RollbackStorage(ctx context.Context, storageID, snapshotID string, body gsclient.StorageRollbackRequest) error {
	panic("implement me")
}

func (s SnapshotOperatorMock) ExportStorageSnapshotToS3(ctx context.Context, storageID, snapshotID string, body gsclient.StorageSnapshotExportToS3Request) error {
	panic("implement me")
}

func Test_stepCreateSnapshot_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.StorageSnapshotOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		state multistep.StateBag
	}
	ui := &uiMock{}
	testConfig := produceTestConfig(make(map[string]interface{}))
	tests := []struct {
		name    string
		fields  fields
		args    args
		success bool
		message string
	}{
		{
			name: "success",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "test UUID",
					"snapshot_uuid":     "success",
				}},
			},
			success: true,
			message: "Destroyed the snapshot (success) of storage (test UUID)",
		},
		{
			name: "API call fail",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "test UUID",
					"snapshot_uuid":     "fail",
				}},
			},
			success: false,
			message: "Error destroying snapshot. Please destroy it manually: error",
		},
		{
			name: "convert boot_storage_uuid to string fail",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert boot_storage_uuid to string",
		},
		{
			name: "empty boot_storage_uuid",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "",
				}},
			},
			success: false,
			message: "boot_storage_uuid is empty",
		},
		{
			name: "convert snapshot_uuid to string fail",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "test UUID",
				}},
			},
			success: false,
			message: "cannot convert snapshot_uuid to string",
		},
		{
			name: "empty snapshot_uuid",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "test UUID",
					"snapshot_uuid":     "",
				}},
			},
			success: false,
			message: "snapshot_uuid is empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateSnapshot{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			s.Cleanup(tt.args.state)
			if tt.success {
				if ui.sayMessage != tt.message {
					t.Errorf("Cleanup() got message = %v, want %v", ui.sayMessage, tt.message)
				}
			} else {
				if ui.errorMessage != tt.message {
					t.Errorf("Cleanup() got error message = %v, want %v", ui.errorMessage, tt.message)
				}
			}
		})
	}
}

func Test_stepCreateSnapshot_Run(t *testing.T) {
	type fields struct {
		client gsclient.StorageSnapshotOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		ctx   context.Context
		state multistep.StateBag
	}
	ui := &uiMock{}
	testConfig := produceTestConfig(make(map[string]interface{}))
	tests := []struct {
		name   string
		fields fields
		args   args
		want   multistep.StepAction
	}{
		{
			name: "success",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "success",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert boot_storage_uuid to string fail",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx:   context.Background(),
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty boot_storage_uuid",
			fields: fields{
				client: SnapshotOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateSnapshot{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				uuid, ok := tt.args.state.Get("snapshot_uuid").(string)
				if !ok {
					t.Error("cannot convert snapshot_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("snapshot_uuid = %v, want test", uuid)
				}
			}
		})
	}
}
