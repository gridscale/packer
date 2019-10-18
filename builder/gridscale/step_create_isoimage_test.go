package gridscale

import (
	"context"
	"errors"
	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"testing"
)

type ISOImageOperatorMock struct{}

func (I ISOImageOperatorMock) GetISOImageList(ctx context.Context) ([]gsclient.ISOImage, error) {
	panic("implement me")
}

func (I ISOImageOperatorMock) GetISOImage(ctx context.Context, id string) (gsclient.ISOImage, error) {
	panic("implement me")
}

func (I ISOImageOperatorMock) CreateISOImage(ctx context.Context, body gsclient.ISOImageCreateRequest) (gsclient.ISOImageCreateResponse, error) {
	if body.Name == "success" {
		return gsclient.ISOImageCreateResponse{
			ObjectUUID:  "test",
			RequestUUID: "test",
		}, nil
	}
	return gsclient.ISOImageCreateResponse{}, errors.New("error")
}

func (I ISOImageOperatorMock) UpdateISOImage(ctx context.Context, id string, body gsclient.ISOImageUpdateRequest) error {
	panic("implement me")
}

func (I ISOImageOperatorMock) DeleteISOImage(ctx context.Context, id string) error {
	if id == "success" {
		return nil
	}
	return errors.New("error")
}

func (I ISOImageOperatorMock) GetISOImageEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func (I ISOImageOperatorMock) GetISOImagesByLocation(ctx context.Context, id string) ([]gsclient.ISOImage, error) {
	panic("implement me")
}

func (I ISOImageOperatorMock) GetDeletedISOImages(ctx context.Context) ([]gsclient.ISOImage, error) {
	panic("implement me")
}

func Test_stepCreateISOImage_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ISOImageOperator
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
			name: "skip when IsoImageUUID is set",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_uuid": "test UUID",
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: true,
			message: "",
		},
		{
			name: "skip when IsoImageUUID and IsoImageURL are NOT set",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: true,
			message: "No ISO-image is requested. Skipping removing ISO-image...",
		},
		{
			name: "success",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_url": "test URL",
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"iso_image_uuid": "success",
				}},
			},
			success: true,
			message: "Destroyed the ISO-image (success)",
		},
		{
			name: "API call fail",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_url": "test URL",
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"iso_image_uuid": "fail",
				}},
			},
			success: false,
			message: "Error destroying ISO-image (fail). Please destroy it manually: error",
		},
		{
			name: "convert iso_image_uuid to string fail",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_url": "test URL",
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: make(map[string]interface{})},
			},
			success: false,
			message: "cannot convert iso_image_uuid to string",
		},
		{
			name: "no ISO-image UUID detected",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_url": "test URL",
				}),
				ui: ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"iso_image_uuid": "",
				}},
			},
			success: true,
			message: "No ISO-image UUID detected.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepCreateISOImage{
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

func Test_stepCreateISOImage_Run(t *testing.T) {
	type fields struct {
		client gsclient.ISOImageOperator
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
			name: "No ISO-image is requested",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: testConfig,
				ui:     ui,
			},
			args: args{
				ctx:   context.Background(),
				state: StateBagMock{state: make(map[string]interface{})},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "Get ISO-image UUID from config",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_uuid": "test UUID",
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
			name: "Creat an ISO-image success",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_url": "test URL",
					"server_name":  "success",
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
			name: "Creat an ISO-image fail",
			fields: fields{
				client: ISOImageOperatorMock{},
				config: produceTestConfig(map[string]interface{}{
					"isoimage_url": "test URL",
					"server_name":  "fail",
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
			s := &stepCreateISOImage{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue &&
				tt.fields.config.IsoImageUUID == "" &&
				tt.fields.config.IsoImageURL != "" {
				uuid, ok := tt.args.state.Get("iso_image_uuid").(string)
				if !ok {
					t.Error("cannot convert iso_image_uuid to string")
				}
				if uuid != "test" {
					t.Errorf("iso_image_uuid = %v, want test", uuid)
				}
			}
		})
	}
}
