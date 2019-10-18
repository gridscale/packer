package gridscale

import (
	"context"
	"errors"
	"testing"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type ServerIsoImageRelationOperatorMock struct{}

func (s ServerIsoImageRelationOperatorMock) GetServerIsoImageList(ctx context.Context, id string) ([]gsclient.ServerIsoImageRelationProperties, error) {
	panic("implement me")
}

func (s ServerIsoImageRelationOperatorMock) GetServerIsoImage(ctx context.Context, serverID, isoImageID string) (gsclient.ServerIsoImageRelationProperties, error) {
	panic("implement me")
}

func (s ServerIsoImageRelationOperatorMock) CreateServerIsoImage(ctx context.Context, id string, body gsclient.ServerIsoImageRelationCreateRequest) error {
	panic("implement me")
}

func (s ServerIsoImageRelationOperatorMock) UpdateServerIsoImage(ctx context.Context, serverID, isoImageID string, body gsclient.ServerIsoImageRelationUpdateRequest) error {
	panic("implement me")
}

func (s ServerIsoImageRelationOperatorMock) DeleteServerIsoImage(ctx context.Context, serverID, isoImageID string) error {
	panic("implement me")
}

func (s ServerIsoImageRelationOperatorMock) LinkIsoImage(ctx context.Context, serverID string, isoimageID string) error {
	if isoimageID == "success" {
		return nil
	}
	return errors.New("error")
}

func (s ServerIsoImageRelationOperatorMock) UnlinkIsoImage(ctx context.Context, serverID string, isoimageID string) error {
	if isoimageID == "success" {
		return nil
	}
	return errors.New("error")
}

func Test_stepLinkServerISOImage_Cleanup(t *testing.T) {
	type fields struct {
		client gsclient.ServerIsoImageRelationOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		state multistep.StateBag
	}
	ui := &uiMock{}
	testConfigISOImage := produceTestConfig(map[string]interface{}{
		"isoimage_uuid": "test",
	})
	testConfigNoISOImage := produceTestConfig(make(map[string]interface{}))
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
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"iso_image_uuid":          "success",
					"server_iso_image_linked": true,
				}},
			},
			success: true,
			message: "Unlinked the server (test UUID) and the ISO-image (success)",
		},
		{
			name: "skip due to no ISO-Image",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigNoISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"iso_image_uuid":          "success",
					"server_iso_image_linked": true,
				}},
			},
			success: true,
			message: "No ISO-image is requested. Skipping unlinking server with an ISO-image...",
		},
		{
			name: "HTTP call fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"iso_image_uuid":          "fail",
					"server_iso_image_linked": true,
				}},
			},
			success: false,
			message: "Error unlink server (test UUID) and ISO-image (fail). Please unlink them manually: error",
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"iso_image_uuid":          "success",
					"server_iso_image_linked": true,
				}},
			},
			success: false,
			message: "cannot convert server_uuid to string",
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "",
					"iso_image_uuid":          "success",
					"server_iso_image_linked": true,
				}},
			},
			success: false,
			message: "serverUUID is empty",
		},
		{
			name: "convert iso_image_uuid to string fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"server_iso_image_linked": true,
				}},
			},
			success: false,
			message: "cannot convert iso_image_uuid to string",
		},
		{
			name: "empty iso_image_uuid",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"iso_image_uuid":          "",
					"server_iso_image_linked": true,
				}},
			},
			success: true,
			message: "No ISO-image UUID detected. Skipping unlinking the server with an ISO-image...",
		},
		{
			name: "convert server_iso_image_linked to boolean fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":    "test UUID",
					"iso_image_uuid": "success",
				}},
			},
			success: false,
			message: "cannot convert server_iso_image_linked to boolean",
		},
		{
			name: "server_iso_image_linked is false",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"iso_image_uuid":          "success",
					"server_iso_image_linked": false,
				}},
			},
			success: true,
			message: "the server is not linked with the ISO-image.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerISOImage{
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

func Test_stepLinkServerISOImage_Run(t *testing.T) {
	type fields struct {
		client gsclient.ServerIsoImageRelationOperator
		config *Config
		ui     packer.Ui
	}
	type args struct {
		ctx   context.Context
		state multistep.StateBag
	}
	ui := &uiMock{}
	testConfigISOImage := produceTestConfig(map[string]interface{}{
		"isoimage_uuid": "test",
	})
	testConfigNoISOImage := produceTestConfig(make(map[string]interface{}))
	tests := []struct {
		name   string
		fields fields
		args   args
		want   multistep.StepAction
	}{
		{
			name: "success",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":    "test UUID",
					"iso_image_uuid": "success",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "skip due to no ISO-Image",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigNoISOImage,
				ui:     ui,
			},
			args: args{
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":             "test UUID",
					"iso_image_uuid":          "success",
					"server_iso_image_linked": true,
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "API call fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":    "test UUID",
					"iso_image_uuid": "fail",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "convert iso_image_uuid to string fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
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
			name: "empty iso_image_uuid",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":    "test UUID",
					"iso_image_uuid": "",
				}},
			},
			want: multistep.ActionContinue,
		},
		{
			name: "convert server_uuid to string fail",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"iso_image_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
		{
			name: "empty server_uuid",
			fields: fields{
				client: ServerIsoImageRelationOperatorMock{},
				config: testConfigISOImage,
				ui:     ui,
			},
			args: args{
				ctx: context.Background(),
				state: StateBagMock{state: map[string]interface{}{
					"server_uuid":    "",
					"iso_image_uuid": "success",
				}},
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stepLinkServerISOImage{
				client: tt.fields.client,
				config: tt.fields.config,
				ui:     tt.fields.ui,
			}
			if got := s.Run(tt.args.ctx, tt.args.state); got != tt.want {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
			if tt.want == multistep.ActionContinue &&
				tt.args.state.Get("iso_image_uuid").(string) != "" &&
				(tt.fields.config.IsoImageUUID != "" ||
					tt.fields.config.IsoImageURL != "") {
				linked, ok := tt.args.state.Get("server_iso_image_linked").(bool)
				if !ok {
					t.Error("cannot convert server_iso_image_linked to boolean")
				}
				if !linked {
					t.Errorf("server_iso_image_linked = %v, want true", linked)
				}
			}
		})
	}
}
