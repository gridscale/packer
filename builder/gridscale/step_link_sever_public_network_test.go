package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type ServerNetworkRelationOperatorMock struct{}

func (s ServerNetworkRelationOperatorMock) GetServerNetworkList(ctx context.Context, id string) ([]gsclient.ServerNetworkRelationProperties, error) {
	panic("implement me")
}

func (s ServerNetworkRelationOperatorMock) GetServerNetwork(ctx context.Context, serverID, networkID string) (gsclient.ServerNetworkRelationProperties, error) {
	panic("implement me")
}

func (s ServerNetworkRelationOperatorMock) CreateServerNetwork(ctx context.Context, id string, body gsclient.ServerNetworkRelationCreateRequest) error {
	panic("implement me")
}

func (s ServerNetworkRelationOperatorMock) UpdateServerNetwork(ctx context.Context, serverID, networkID string, body gsclient.ServerNetworkRelationUpdateRequest) error {
	panic("implement me")
}

func (s ServerNetworkRelationOperatorMock) DeleteServerNetwork(ctx context.Context, serverID, networkID string) error {
	panic("implement me")
}

func (s ServerNetworkRelationOperatorMock) LinkNetwork(ctx context.Context, serverID, networkID, firewallTemplate string, bootdevice bool, order int, l3security []string, firewall *gsclient.FirewallRules) error {
	if networkID == "success" {
		return nil
	}
	return errors.New("error")
}

func (s ServerNetworkRelationOperatorMock) UnlinkNetwork(ctx context.Context, serverID string, networkID string) error {
	if networkID == "success" {
		return nil
	}
	return errors.New("error")
}

func Test_stepLinkServerPublicNetwork_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerNetworkRelationOperator
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
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                  "test UUID",
					"public_network_uuid":          "success",
					"server_public_network_linked": true,
				}},
			},
			success: true,
			message: "Unlinked the server (test UUID) and the public network (success)",
		},
		{
			name: "HTTP call fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                  "test UUID",
					"public_network_uuid":          "fail",
					"server_public_network_linked": true,
				}},
			},
			success: false,
			message: "Error unlink server (test UUID) and public network (fail). Please unlink them manually: error",
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"public_network_uuid":          "success",
					"server_public_network_linked": true,
				}},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                  "",
					"public_network_uuid":          "success",
					"server_public_network_linked": true,
				}},
			},
			success: false,
			message: "serverUUID is empty",
		},
		{
			name: "convert public_network_uuid to string fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                  "test UUID",
					"server_public_network_linked": true,
				}},
			},
			success: false,
			message: "cannot convert public_network_uuid to string",
		},
		{
			name: "empty public_network_uuid",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                  "test UUID",
					"public_network_uuid":          "",
					"server_public_network_linked": true,
				}},
			},
			success: false,
			message: "public_network_uuid is empty",
		},
		{
			name: "convert server_public_network_linked to boolean fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":         "test UUID",
					"public_network_uuid": "success",
				}},
			},
			success: false,
			message: "cannot convert server_public_network_linked to boolean",
		},
		{
			name: "server_public_network_linked is false",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":                  "test UUID",
					"public_network_uuid":          "success",
					"server_public_network_linked": false,
				}},
			},
			success: true,
			message: "the server is not linked with the public network.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerPublicNetwork{
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

func Test_stepLinkServerPublicNetwork_Run(t *testing.T) {
	type fields struct {
		client gsclient.ServerNetworkRelationOperator
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
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":         "test UUID",
					"public_network_uuid": "success",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":         "test UUID",
					"public_network_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert public_network_uuid to string fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
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
			name: "empty public_network_uuid",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":         "test UUID",
					"public_network_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"public_network_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerNetworkRelationOperatorMock{},
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":         "",
					"public_network_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerPublicNetwork{
				client: tt.fields.client,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				linked, ok := tt.args.state.Get("server_public_network_linked").(bool)
				if !ok {
					t.Error("cannot convert server_public_network_linked to boolean")
				}
				if !linked {
					t.Errorf("server_public_network_linked = %v, want true", linked)
				}
			}
		})
	}
}
