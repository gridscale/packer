package gridscale

import (
	"context"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func Test_stepCreateSecondaryStorage_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.StorageOperator
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
			name: "skip",
			fields: fields{
				client: StorageOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: true,
			message: "",
		},
		{
			name: "success",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"secondary_storage": true,
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"secondary_storage_uuid": "success",
				}},
			},
			success: true,
			message: "Destroyed the secondary storage (success)",
		},
		{
			name: "API call fail",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"secondary_storage": true,
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"secondary_storage_uuid": "fail",
				}},
			},
			success: false,
			message: "Error destroying secondary storage (fail). Please destroy it manually: error",
		},
		{
			name: "convert secondary_storage_uuid to string error",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"secondary_storage": true,
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert secondary_storage_uuid to string",
		},
		{
			name: "No secondary storage UUID detected",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"secondary_storage": true,
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"secondary_storage_uuid": "",
				}},
			},
			success: true,
			message: "No secondary storage UUID detected.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateSecondaryStorage{
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

func Test_stepCreateSecondaryStorage_Run(t *testing.T) {
	type fields struct {
		client gsclient.StorageOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		ctx   context.Context
		state multistep.StateBag
	}
	ui := &uiMock{}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   multistep.StepAction
	}{
		{
			name: "skip",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(make(map[string]interface{})),
				ui:     ui,
			},
			args: args{
				ctx:   context.Background(),
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "success",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"server_name":       "success",
					"secondary_storage": true,
				}),
				ui: ui,
			},
			args: args{
				ctx:   context.Background(),
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: StorageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"server_name":       "fail",
					"secondary_storage": true,
				}),
				ui: ui,
			},
			args: args{
				ctx:   context.Background(),
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateSecondaryStorage{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("stepCreateBootStorage_Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue && tt.fields.config.SecondaryStorage {
				uuid, ok := tt.args.state.Get("secondary_storage_uuid").(string)
				if !ok {
					t.Error("cannot convert secondary_storage_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("secondary_storage_uuid = %v, want test", uuid)
				}
			}
		})
	}
}
