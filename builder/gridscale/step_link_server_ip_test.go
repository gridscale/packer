package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type ServerIPRelationOperatorMock struct{}

func (s ServerIPRelationOperatorMock) GetServerIPList(ctx context.Context, id string) ([]gsclient.ServerIPRelationProperties, error) {
	panic("implement me")
}

func (s ServerIPRelationOperatorMock) GetServerIP(ctx context.Context, serverID, ipID string) (gsclient.ServerIPRelationProperties, error) {
	panic("implement me")
}

func (s ServerIPRelationOperatorMock) CreateServerIP(ctx context.Context, id string, body gsclient.ServerIPRelationCreateRequest) error {
	panic("implement me")
}

func (s ServerIPRelationOperatorMock) DeleteServerIP(ctx context.Context, serverID, ipID string) error {
	panic("implement me")
}

func (s ServerIPRelationOperatorMock) LinkIP(ctx context.Context, serverID string, ipID string) error {
	if ipID == "success" {
		return nil
	}
	return errors.New("error")
}

func (s ServerIPRelationOperatorMock) UnlinkIP(ctx context.Context, serverID string, ipID string) error {
	if ipID == "success" {
		return nil
	}
	return errors.New("error")
}

func Test_stepLinkServerIPAddr_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerIPRelationOperator
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
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":           "test UUID",
					"ip_addr_uuid":          "success",
					"server_ip_addr_linked": true,
				}},
			},
			success: true,
			message: "Unlinked the server (test UUID) and the IP address (success)",
		},
		{
			name: "HTTP call fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":           "test UUID",
					"ip_addr_uuid":          "fail",
					"server_ip_addr_linked": true,
				}},
			},
			success: false,
			message: "Error unlink server (test UUID) and IP address (fail). Please unlink them manually: error",
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ip_addr_uuid":          "success",
					"server_ip_addr_linked": true,
				}},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":           "",
					"ip_addr_uuid":          "success",
					"server_ip_addr_linked": true,
				}},
			},
			success: false,
			message: "serverUUID is empty",
		},
		{
			name: "convert ip_addr_uuid to string fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":           "test UUID",
					"server_ip_addr_linked": true,
				}},
			},
			success: false,
			message: "cannot convert ip_addr_uuid to string",
		},
		{
			name: "empty ip_addr_uuid",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":           "test UUID",
					"ip_addr_uuid":          "",
					"server_ip_addr_linked": true,
				}},
			},
			success: false,
			message: "ip_addr_uuid is empty",
		},
		{
			name: "convert server_ip_addr_linked to boolean fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":  "test UUID",
					"ip_addr_uuid": "success",
				}},
			},
			success: false,
			message: "cannot convert server_ip_addr_linked to boolean",
		},
		{
			name: "server_ip_addr_linked is false",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":           "test UUID",
					"ip_addr_uuid":          "success",
					"server_ip_addr_linked": false,
				}},
			},
			success: true,
			message: "the server is not linked with the IP address.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerIPAddr{
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

func Test_stepLinkServerIPAddr_Run(t *testing.T) {
	type fields struct {
		client gsclient.ServerIPRelationOperator
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
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":  "test UUID",
					"ip_addr_uuid": "success",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":  "test UUID",
					"ip_addr_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert ip_addr_uuid to string fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
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
			name: "empty ip_addr_uuid",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":  "test UUID",
					"ip_addr_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"ip_addr_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerIPRelationOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":  "",
					"ip_addr_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerIPAddr{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				linked, ok := tt.args.state.Get("server_ip_addr_linked").(bool)
				if !ok {
					t.Error("cannot convert server_ip_addr_linked to boolean")
				}
				if !linked {
					t.Errorf("server_ip_addr_linked = %v, want true", linked)
				}
			}
		})
	}
}
