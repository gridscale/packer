package gridscale

import (
	"context"
	"errors"
	"github.com/gridscale/gsclient-go/v3"
	"reflect"
	"testing"
)

type TemplateOperatorMock struct{}

func (t TemplateOperatorMock) GetTemplate(ctx context.Context, id string) (gsclient.Template, error) {
	panic("implement me")
}

func (t TemplateOperatorMock) GetTemplateByName(ctx context.Context, name string) (gsclient.Template, error) {
	panic("implement me")
}

func (t TemplateOperatorMock) GetTemplateList(ctx context.Context) ([]gsclient.Template, error) {
	panic("implement me")
}

func (t TemplateOperatorMock) CreateTemplate(ctx context.Context, body gsclient.TemplateCreateRequest) (gsclient.CreateResponse, error) {
	if body.Name == "success" {
		return gsclient.CreateResponse{
			ObjectUUID:  "test",
			RequestUUID: "test",
		}, nil
	}
	return gsclient.CreateResponse{}, errors.New("error")
}

func (t TemplateOperatorMock) UpdateTemplate(ctx context.Context, id string, body gsclient.TemplateUpdateRequest) error {
	panic("implement me")
}

func (t TemplateOperatorMock) DeleteTemplate(ctx context.Context, id string) error {
	if id == "success" {
		return nil
	}
	return errors.New("error")
}

func (t TemplateOperatorMock) GetDeletedTemplates(ctx context.Context) ([]gsclient.Template, error) {
	panic("implement me")
}

func (t TemplateOperatorMock) GetTemplateEventList(ctx context.Context, id string) ([]gsclient.Event, error) {
	panic("implement me")
}

func TestArtifact_BuilderId(t *testing.T) {
	type fields struct {
		TemplateName string
		TemplateUUID string
		Client       gsclient.TemplateOperator
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "get builder ID",
			fields: fields{
				TemplateName: "test",
				TemplateUUID: "test UUID",
				Client:       nil,
			},
			want: BuilderId,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := &Artifact{
				TemplateName: tt.fields.TemplateName,
				TemplateUUID: tt.fields.TemplateUUID,
				Client:       tt.fields.Client,
			}
			if got := ar.BuilderId(); got != tt.want {
				t.Errorf("BuilderId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtifact_Destroy(t *testing.T) {
	type fields struct {
		TemplateName string
		TemplateUUID string
		Client       gsclient.TemplateOperator
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "destroy template success",
			fields: fields{
				TemplateName: "test",
				TemplateUUID: "success",
				Client:       TemplateOperatorMock{},
			},
			wantErr: false,
		},
		{
			name: "destroy template fail",
			fields: fields{
				TemplateName: "test",
				TemplateUUID: "fail",
				Client:       TemplateOperatorMock{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Artifact{
				TemplateName: tt.fields.TemplateName,
				TemplateUUID: tt.fields.TemplateUUID,
				Client:       tt.fields.Client,
			}
			if err := a.Destroy(); (err != nil) != tt.wantErr {
				t.Errorf("Destroy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArtifact_Files(t *testing.T) {
	type fields struct {
		TemplateName string
		TemplateUUID string
		Client       gsclient.TemplateOperator
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name:   "nil artifact file",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := &Artifact{
				TemplateName: tt.fields.TemplateName,
				TemplateUUID: tt.fields.TemplateUUID,
				Client:       tt.fields.Client,
			}
			if got := ar.Files(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Files() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtifact_Id(t *testing.T) {
	type fields struct {
		TemplateName string
		TemplateUUID string
		Client       gsclient.TemplateOperator
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Get artifact ID",
			fields: fields{
				TemplateName: "test",
				TemplateUUID: "test UUID",
				Client:       nil,
			},
			want: "test UUID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Artifact{
				TemplateName: tt.fields.TemplateName,
				TemplateUUID: tt.fields.TemplateUUID,
				Client:       tt.fields.Client,
			}
			if got := a.Id(); got != tt.want {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtifact_State(t *testing.T) {
	type fields struct {
		TemplateName string
		TemplateUUID string
		Client       gsclient.TemplateOperator
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{
			name:   "Get artifact state",
			fields: fields{},
			args:   args{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Artifact{
				TemplateName: tt.fields.TemplateName,
				TemplateUUID: tt.fields.TemplateUUID,
				Client:       tt.fields.Client,
			}
			if got := a.State(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("State() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtifact_String(t *testing.T) {
	type fields struct {
		TemplateName string
		TemplateUUID string
		Client       gsclient.TemplateOperator
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Get artifact string",
			fields: fields{
				TemplateName: "test",
				TemplateUUID: "test UUID",
				Client:       nil,
			},
			want: "A template was created: 'test' (ID: test UUID) ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Artifact{
				TemplateName: tt.fields.TemplateName,
				TemplateUUID: tt.fields.TemplateUUID,
				Client:       tt.fields.Client,
			}
			if got := a.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
