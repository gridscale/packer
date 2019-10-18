package gridscale

import (
	"context"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func Test_stepLinkServerSecondaryStorage_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerStorageRelationOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		state multistep.StateBag
	}
	ui := &uiMock{}
	testConfigSecondaryStorage := produceTestConfig(map[string]interface{}{
		"secondary_storage": true,
	})
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
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"secondary_storage_uuid":          "success",
					"server_secondary_storage_linked": true,
				}},
			},
			success: true,
			message: "Unlinked the server (test UUID) and the secondary storage (success)",
		},
		{
			name: "skip due to no secondary storage",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: produceTestConfig(make(map[string]interface{})),
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"secondary_storage_uuid":          "success",
					"server_secondary_storage_linked": true,
				}},
			},
			success: true,
			message: "No secondary storage is requested. Skipping unlinking server with secondary storage...",
		},
		{
			name: "HTTP call fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"secondary_storage_uuid":          "fail",
					"server_secondary_storage_linked": true,
				}},
			},
			success: false,
			message: "Error unlink server (test UUID) and secondary storage (fail). Please unlink them manually: error",
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"secondary_storage_uuid":          "success",
					"server_secondary_storage_linked": true,
				}},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "",
					"secondary_storage_uuid":          "success",
					"server_secondary_storage_linked": true,
				}},
			},
			success: false,
			message: "serverUUID is empty",
		},
		{
			name: "convert secondary_storage_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"server_secondary_storage_linked": true,
				}},
			},
			success: false,
			message: "cannot convert secondary_storage_uuid to string",
		},
		{
			name: "empty secondary_storage_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"secondary_storage_uuid":          "",
					"server_secondary_storage_linked": true,
				}},
			},
			success: false,
			message: "secondary_storage_uuid is empty",
		},
		{
			name: "convert server_secondary_storage_linked to boolean fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":            "test UUID",
					"secondary_storage_uuid": "success",
				}},
			},
			success: false,
			message: "cannot convert server_secondary_storage_linked to boolean",
		},
		{
			name: "server_secondary_storage_linked is false",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"secondary_storage_uuid":          "success",
					"server_secondary_storage_linked": false,
				}},
			},
			success: true,
			message: "the server is not linked with the secondary storage.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerSecondaryStorage{
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

func Test_stepLinkServerSecondaryStorage_Run(t *testing.T) {
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
	testConfigSecondaryStorage := produceTestConfig(map[string]interface{}{
		"secondary_storage": true,
	})
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
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":            "test UUID",
					"secondary_storage_uuid": "success",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "skip due to no secondary storage",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: produceTestConfig(make(map[string]interface{})),
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                     "test UUID",
					"secondary_storage_uuid":          "success",
					"server_secondary_storage_linked": true,
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":            "test UUID",
					"secondary_storage_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert secondary_storage_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
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
			name: "empty secondary_storage_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":            "test UUID",
					"secondary_storage_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"secondary_storage_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerStorageRelationOperatorMock{},
				config: testConfigSecondaryStorage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":            "",
					"secondary_storage_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerSecondaryStorage{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue && tt.fields.config.SecondaryStorage {
				linked, ok := tt.args.state.Get("server_secondary_storage_linked").(bool)
				if !ok {
					t.Error("cannot convert server_secondary_storage_linked to boolean")
				}
				if !linked {
					t.Errorf("server_secondary_storage_linked = %v, want true", linked)
				}
			}
		})
	}
}
