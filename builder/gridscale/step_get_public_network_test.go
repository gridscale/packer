package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type NetworkOperatorMock struct {
	getNetworkPublicSuccess bool
}

func (n NetworkOperatorMock) GetNetwork(ctx context.Context, id string) (gsclient.Network, error) {
	panic("implement me")
}

func (n NetworkOperatorMock) GetNetworkList(ctx context.Context) ([]gsclient.Network, error) {
	panic("implement me")
}

func (n NetworkOperatorMock) CreateNetwork(ctx context.Context, body gsclient.NetworkCreateRequest) (gsclient.NetworkCreateResponse, error) {
	panic("implement me")
}

func (n NetworkOperatorMock) DeleteNetwork(ctx context.Context, id string) error {
	panic("implement me")
}

func (n NetworkOperatorMock) UpdateNetwork(ctx context.Context, id string, body gsclient.NetworkUpdateRequest) error {
	panic("implement me")
}

func (n NetworkOperatorMock) GetNetworkEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func (n NetworkOperatorMock) GetNetworkPublic(ctx context.Context) (gsclient.Network, error) {
	if n.getNetworkPublicSuccess {
		return gsclient.Network{Properties: gsclient.NetworkProperties{ObjectUUID: "test UUID"}}, nil
	}
	return gsclient.Network{}, errors.New("error")
}

func (n NetworkOperatorMock) GetNetworksByLocation(ctx context.Context, id string) ([]gsclient.Network, error) {
	panic("implement me")
}

func (n NetworkOperatorMock) GetDeletedNetworks(ctx context.Context) ([]gsclient.Network, error) {
	panic("implement me")
}

func Test_stepGetPublicNetwork_Cleanup(t *testing.T) {
}

func Test_stepGetPublicNetwork_Run(t *testing.T) {
	type fields struct {
		client gsclient.NetworkOperator
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
				client: NetworkOperatorMock{getNetworkPublicSuccess: true},
				ui:     ui,
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
				client: NetworkOperatorMock{getNetworkPublicSuccess: false},
				ui:     ui,
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
			s := &stepGetPublicNetwork{
				client: tt.fields.client,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				uuid, ok := tt.args.state.Get("public_network_uuid").(string)
				if !ok {
					t.Error("cannot convert public_network_uuid to string")
				}
				if uuid != "test UUID" {
					t.Errorf("public_network_uuid = %v, want test UUID", uuid)
				}
			}
		})
	}
}
