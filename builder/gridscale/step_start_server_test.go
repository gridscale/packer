package gridscale

import (
	"context"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func Test_stepStartServer_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerOperator
		ui     packer.Ui
	}
	type args struct {
		state multistep.StateBag
	}
	ui := &uiMock{}
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
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "ShutdownSuccess",
				}},
			},
			success: true,
			message: "Gracefully shut down server (ShutdownSuccess)",
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "fail",
				}},
			},
			success: false,
			message: "Error shutting down server: error",
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "",
				}},
			},
			success: false,
			message: "server_uuid is empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepStartServer{
				client: tt.fields.client,
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

func Test_stepStartServer_Run(t *testing.T) {
	type fields struct {
		client gsclient.ServerOperator
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
			name: "success",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "StartSuccess",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepStartServer{
				client: tt.fields.client,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
