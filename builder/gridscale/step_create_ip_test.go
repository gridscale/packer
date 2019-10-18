package gridscale

import (
	"context"
	"errors"
	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	"testing"
)

type IPOperatorMock struct{}

func (I IPOperatorMock) GetIP(ctx context.Context, id string) (gsclient.IP, error) {
	panic("implement me")
}

func (I IPOperatorMock) GetIPList(ctx context.Context) ([]gsclient.IP, error) {
	panic("implement me")
}

func (I IPOperatorMock) CreateIP(ctx context.Context, body gsclient.IPCreateRequest) (gsclient.IPCreateResponse, error) {
	if body.Name == "success" {
		return gsclient.IPCreateResponse{
			ObjectUUID:  "test",
			RequestUUID: "test",
			IP:          "test",
		}, nil
	}
	return gsclient.IPCreateResponse{}, errors.New("error")
}

func (I IPOperatorMock) DeleteIP(ctx context.Context, id string) error {
	if id == "success" {
		return nil
	}
	return errors.New("error")
}

func (I IPOperatorMock) UpdateIP(ctx context.Context, id string, body gsclient.IPUpdateRequest) error {
	panic("implement me")
}

func (I IPOperatorMock) GetIPEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func (I IPOperatorMock) GetIPVersion(ctx context.Context, id string) int {
	panic("implement me")
}

func (I IPOperatorMock) GetIPsByLocation(ctx context.Context, id string) ([]gsclient.IP, error) {
	panic("implement me")
}

func (I IPOperatorMock) GetDeletedIPs(ctx context.Context) ([]gsclient.IP, error) {
	panic("implement me")
}

func Test_stepCreateIPAddr_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.IPOperator
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
				client: IPOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ip_addr_uuid": "success",
				}},
			},
			success: true,
			message: "Destroyed the IP address (success)",
		},
		{
			name: "API call fail",
			fields: fields{
				client: IPOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ip_addr_uuid": "fail",
				}},
			},
			success: false,
			message: "Error destroying IP address (fail). Please destroy it manually: error",
		},
		{
			name: "convert ip_addr_uuid to string error",
			fields: fields{
				client: IPOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert ip_addr_uuid to string",
		},
		{
			name: "No IP address detected",
			fields: fields{
				client: IPOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ip_addr_uuid": "",
				}},
			},
			success: true,
			message: "No IP address UUID detected.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateIPAddr{
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

func Test_stepCreateIPAddr_Run(t *testing.T) {
	type fields struct {
		client gsclient.IPOperator
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
			name: "success",
			fields: fields{
				client: IPOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"server_name": "success",
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
				client: IPOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"server_name": "fail",
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
			s := &stepCreateIPAddr{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				uuid, ok := tt.args.state.Get("ip_addr_uuid").(string)
				if !ok {
					t.Error("cannot convert ip_addr_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("ip_addr_uuid = %v, want test", uuid)
				}
				ip, ok := tt.args.state.Get("server_ip").(string)
				if !ok {
					t.Error("cannot convert server_ip to string")
				}
				if ip != "test" {
					t.Errorf("server_ip = %v, want test", uuid)
				}
			}
		})
	}
}
