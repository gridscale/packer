package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type ServerStorageRelationOperatorMock struct{}

func (s ServerStorageRelationOperatorMock) GetServerStorageList(ctx context.Context, id string) ([]gsclient.ServerStorageRelationProperties, error) {
	panic("implement me")
}

func (s ServerStorageRelationOperatorMock) GetServerStorage(ctx context.Context, serverID, storageID string) (gsclient.ServerStorageRelationProperties, error) {
	panic("implement me")
}

func (s ServerStorageRelationOperatorMock) CreateServerStorage(ctx context.Context, id string, body gsclient.ServerStorageRelationCreateRequest) error {
	panic("implement me")
}

func (s ServerStorageRelationOperatorMock) UpdateServerStorage(ctx context.Context, serverID, storageID string, body gsclient.ServerStorageRelationUpdateRequest) error {
	panic("implement me")
}

func (s ServerStorageRelationOperatorMock) DeleteServerStorage(ctx context.Context, serverID, storageID string) error {
	panic("implement me")
}

func (s ServerStorageRelationOperatorMock) LinkStorage(ctx context.Context, serverID string, storageID string, bootdevice bool) error {
	if storageID == "success" {
		return nil
	}
	return errors.New("error")
}

func (s ServerStorageRelationOperatorMock) UnlinkStorage(ctx context.Context, serverID string, storageID string) error {
	if storageID == "success" {
		return nil
	}
	return errors.New("error")
}

func Test_stepLinkServerBootStorage_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerStorageRelationOperator
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
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                "test UUID",
					"boot_storage_uuid":          "success",
					"server_boot_storage_linked": true,
				}},
			},
			success: true,
			message: "Unlinked the server (test UUID) and the boot storage (success)",
		},
		{
			name: "HTTP call fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                "test UUID",
					"boot_storage_uuid":          "fail",
					"server_boot_storage_linked": true,
				}},
			},
			success: false,
			message: "Error unlink server (test UUID) and boot storage (fail). Please unlink them manually: error",
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid":          "success",
					"server_boot_storage_linked": true,
				}},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                "",
					"boot_storage_uuid":          "success",
					"server_boot_storage_linked": true,
				}},
			},
			success: false,
			message: "serverUUID is empty",
		},
		{
			name: "convert boot_storage_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                "test UUID",
					"server_boot_storage_linked": true,
				}},
			},
			success: false,
			message: "cannot convert boot_storage_uuid to string",
		},
		{
			name: "empty boot_storage_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                "test UUID",
					"boot_storage_uuid":          "",
					"server_boot_storage_linked": true,
				}},
			},
			success: false,
			message: "boot_storage_uuid is empty",
		},
		{
			name: "convert server_boot_storage_linked to boolean fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":       "test UUID",
					"boot_storage_uuid": "success",
				}},
			},
			success: false,
			message: "cannot convert server_boot_storage_linked to boolean",
		},
		{
			name: "server_boot_storage_linked is false",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                "test UUID",
					"boot_storage_uuid":          "success",
					"server_boot_storage_linked": false,
				}},
			},
			success: true,
			message: "the server is not linked with the boot storage.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerBootStorage{
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

func Test_stepLinkServerBootStorage_Run(t *testing.T) {
	type fields struct {
		client gsclient.ServerStorageRelationOperator
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
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":       "test UUID",
					"boot_storage_uuid": "success",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":       "test UUID",
					"boot_storage_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert boot_storage_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "test UUID",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty boot_storage_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":       "test UUID",
					"boot_storage_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"boot_storage_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":       "",
					"boot_storage_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerBootStorage{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				linked, ok := tt.args.state.Get("server_boot_storage_linked").(bool)
				if !ok {
					t.Error("cannot convert server_boot_storage_linked to boolean")
				}
				if !linked {
					t.Errorf("server_boot_storage_linked = %v, want true", linked)
				}
			}
		})
	}
}
