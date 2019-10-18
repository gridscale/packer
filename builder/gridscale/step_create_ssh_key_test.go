package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type SSHOperatorMock struct {
	createSSHKeySuccess bool
}

func (S SSHOperatorMock) GetSshkey(ctx context.Context, id string) (gsclient.Sshkey, error) {
	panic("implement me")
}

func (S SSHOperatorMock) GetSshkeyList(ctx context.Context) ([]gsclient.Sshkey, error) {
	panic("implement me")
}

func (S SSHOperatorMock) CreateSshkey(ctx context.Context, body gsclient.SshkeyCreateRequest) (gsclient.CreateResponse, error) {
	if S.createSSHKeySuccess {
		return gsclient.CreateResponse{
			ObjectUUID:  "test",
			RequestUUID: "test",
		}, nil
	}
	return gsclient.CreateResponse{}, errors.New("error")
}

func (S SSHOperatorMock) DeleteSshkey(ctx context.Context, id string) error {
	if id == "success" {
		return nil
	}
	return errors.New("error")
}

func (S SSHOperatorMock) UpdateSshkey(ctx context.Context, id string, body gsclient.SshkeyUpdateRequest) error {
	panic("implement me")
}

func (S SSHOperatorMock) GetSshkeyEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func Test_stepCreateSSHKey_Cleanup(t *testing.T) {
	type fields struct {
		Debug        bool
		DebugKeyPath string
		client       gsclient.SSHKeyOperator
		config       *Config
		ui           packer.Ui
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
				client: SSHOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ssh_key_uuid": "success",
				}},
			},
			success: true,
			message: "Destroyed the SSH-key (success)",
		},
		{
			name: "API call fail",
			fields: fields{
				client: SSHOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ssh_key_uuid": "fail",
				}},
			},
			success: false,
			message: "Error destroying SSH key (fail). Please destroy it manually: error",
		},
		{
			name: "convert ssh_key_uuid to string error",
			fields: fields{
				client: SSHOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert ssh_key_uuid to string",
		},
		{
			name: "empty ssh_key_uuid",
			fields: fields{
				client: SSHOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"ssh_key_uuid": "",
				}},
			},
			success: true,
			message: "No SSH key UUID detected.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateSSHKey{
				Debug:        tt.fields.Debug,
				DebugKeyPath: tt.fields.DebugKeyPath,
				client:       tt.fields.client,
				config:       tt.fields.config,
				ui:           tt.fields.ui,
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

func Test_stepCreateSSHKey_Run(t *testing.T) {
	type fields struct {
		Debug        bool
		DebugKeyPath string
		client       gsclient.SSHKeyOperator
		config       *Config
		ui           packer.Ui
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
				client: SSHOperatorMock{createSSHKeySuccess: true},
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
			name: "API call fail",
			fields: fields{
				client: SSHOperatorMock{createSSHKeySuccess: false},
				config: produceTestConfig(make(map[string]interface{})),
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
			s := &stepCreateSSHKey{
				Debug:        tt.fields.Debug,
				DebugKeyPath: tt.fields.DebugKeyPath,
				client:       tt.fields.client,
				config:       tt.fields.config,
				ui:           tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				uuid, ok := tt.args.state.Get("ssh_key_uuid").(string)
				if !ok {
					t.Error("cannot convert ssh_key_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("ssh_key_uuid = %v, want test", uuid)
				}
			}
		})
	}
}
