package gridscale

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestBuilder_Prepare(t *testing.T) {
	type fields struct {
		config Config
		runner multistep.Runner
	}
	type args struct {
		raws map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		want1   []string
		wantErr bool
	}{
		{
			name:   "valid config",
			fields: fields{},
			args: args{
				raws: map[string]interface{}{
					"api_token":          "test",
					"api_key":            "test",
					"server_cores":       2,
					"server_memory":      4,
					"storage_capacity":   10,
					"root_template_uuid": "test",
					"ssh_username":       "root",
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: false,
		},
		{
			name:   "bad type",
			fields: fields{},
			args: args{
				raws: map[string]interface{}{
					"api_token":          []string{},
					"api_key":            "test",
					"server_cores":       2,
					"server_memory":      4,
					"storage_capacity":   10,
					"root_template_uuid": "test",
					"ssh_username":       "root",
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name:   "invalid key",
			fields: fields{},
			args: args{
				raws: map[string]interface{}{
					"api_token":          "test",
					"api_key":            "test",
					"server_cores":       2,
					"server_memory":      4,
					"storage_capacity":   10,
					"root_template_uuid": "test",
					"ssh_username":       "root",
					"invalid_key":        "",
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				config: tt.fields.config,
				runner: tt.fields.runner,
			}
			got, got1, err := b.Prepare(tt.args.raws)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Prepare() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Prepare() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
