package gridscale

import (
	"context"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func Test_stepCreateTemplate_Cleanup(t *testing.T) {
}

func Test_stepCreateTemplate_Run(t *testing.T) {
	type fields struct {
		client gsclient.TemplateOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		ctx   context.Context
		state multistep.StateBag
	}
	ui := &uiMock{}
	//testConfig := produceTestConfig(make(map[string]interface{}))
	tests := []struct {
		name   string
		fields fields
		args   args
		want   multistep.StepAction
	}{
		{
			name: "success",
			fields: fields{
				client: TemplateOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"template_name": "success",
				}),
				ui: ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"snapshot_uuid": "test UUID",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: TemplateOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"template_name": "fail",
				}),
				ui: ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"snapshot_uuid": "test UUID",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert snapshot_uuid to string fail",
			fields: fields{
				client: TemplateOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"template_name": "fail",
				}),
				ui: ui,
			},
			args: args{
				ctx:   context.Background(),
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty snapshot_uuid",
			fields: fields{
				client: TemplateOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"template_name": "fail",
				}),
				ui: ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"snapshot_uuid": "",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateTemplate{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue {
				uuid, ok := tt.args.state.Get("template_uuid").(string)
				if !ok {
					t.Error("cannot convert template_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("template_uuid = %v, want test", uuid)
				}
			}
		})
	}
}
