package gridscale

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type ServerOperatorMock struct{}

func (s ServerOperatorMock) GetServer(ctx context.Context, id string) (gsclient.Server, error) {
	panic("implement me")
}

func (s ServerOperatorMock) GetServerList(ctx context.Context) ([]gsclient.Server, error) {
	panic("implement me")
}

func (s ServerOperatorMock) GetServersByLocation(ctx context.Context, id string) ([]gsclient.Server, error) {
	panic("implement me")
}

func (s ServerOperatorMock) CreateServer(ctx context.Context, body gsclient.ServerCreateRequest) (gsclient.ServerCreateResponse, error) {
	if strings.Contains(body.Name, "CreateSuccess") {
		return gsclient.ServerCreateResponse{
			ObjectUUID:  "test",
			RequestUUID: "test",
		}, nil
	}
	return gsclient.ServerCreateResponse{}, errors.New("error")
}

func (s ServerOperatorMock) UpdateServer(ctx context.Context, id string, body gsclient.ServerUpdateRequest) error {
	panic("implement me")
}

func (s ServerOperatorMock) DeleteServer(ctx context.Context, id string) error {
	if strings.Contains(id, "DeleteSuccess") {
		return nil
	}
	return errors.New("error")
}

func (s ServerOperatorMock) StartServer(ctx context.Context, id string) error {
	if strings.Contains(id, "StartSuccess") {
		return nil
	}
	return errors.New("error")
}

func (s ServerOperatorMock) StopServer(ctx context.Context, id string) error {
	if strings.Contains(id, "StopSuccess") {
		return nil
	}
	return errors.New("error")
}

func (s ServerOperatorMock) ShutdownServer(ctx context.Context, id string) error {
	if strings.Contains(id, "ShutdownSuccess") {
		return nil
	}
	return errors.New("error")
}

func (s ServerOperatorMock) IsServerOn(ctx context.Context, id string) (bool, error) {
	panic("implement me")
}

func (s ServerOperatorMock) GetServerMetricList(ctx context.Context, id string) ([]gsclient.ServerMetric, error) {
	panic("implement me")
}

func (s ServerOperatorMock) GetServerEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func (s ServerOperatorMock) GetDeletedServers(ctx context.Context) ([]gsclient.Server, error) {
	panic("implement me")
}

func Test_stepCreateServer_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerOperator
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
				client: ServerOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "StopSuccess DeleteSuccess",
				}},
			},
			success: true,
			message: "Destroyed the server (StopSuccess DeleteSuccess)",
		},
		{
			name: "stop server API call fail",
			fields: fields{
				client: ServerOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "fail",
				}},
			},
			success: false,
			message: "Error shutdown server: error",
		},
		{
			name: "delete server API call fail",
			fields: fields{
				client: ServerOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "StopSuccess",
				}},
			},
			success: false,
			message: "Error destroying server. Please destroy it manually: error",
		},
		{
			name: "convert server_uuid to string error",
			fields: fields{
				client: ServerOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "No server detected",
			fields: fields{
				client: ServerOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid": "",
				}},
			},
			success: true,
			message: "No server UUID detected.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateServer{
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

func Test_stepCreateServer_Run(t *testing.T) {
	type fields struct {
		client gsclient.ServerOperator
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
				client: ServerOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"server_name": "CreateSuccess",
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
				client: ServerOperatorMock{},
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
			s := &stepCreateServer{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				uuid, ok := tt.args.state.Get("server_uuid").(string)
				if !ok {
					t.Error("cannot convert server_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("server_uuid = %v, want test", uuid)
				}
			}
		})
	}
}
