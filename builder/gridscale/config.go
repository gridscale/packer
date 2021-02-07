//go:generate mapstructure-to-hcl2 -type Config

package gridscale

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	// The client TOKEN to use to access your account.
	APIToken string `mapstructure:"api_token" required:"true"`
	// The client KEY to use to access your account.
	APIKey string `mapstructure:"api_key" required:"true"`
	// The server URL to use to access your account. Default: "https://api.gridscale.io"
	APIURL string `mapstructure:"api_url" required:"false"`
	// APIRequestHeaders is for debug purpose only. Format: "key1:val1,key2:val2"
	APIRequestHeaders string `mapstructure:"api_request_headers" required:"false"`
	// The name of the new template.
	TemplateName string `mapstructure:"template_name" required:"false"`
	// Name of the host.
	Hostname string `mapstructure:"hostname" required:"false"`
	// Name of the server used for producing the template.
	ServerName string `mapstructure:"server_name" required:"false"`
	// Number of server cores
	ServerCores int `mapstructure:"server_cores" required:"true"`
	// Server memory capacity (in GB)
	ServerMemory int `mapstructure:"server_memory" required:"true"`
	// Storage capacity (in GB)
	StorageCapacity int `mapstructure:"storage_capacity" required:"true"`
	// SecondaryStorage is set to true when the server needs a secondary storage
	// during producing template process.
	SecondaryStorage bool `mapstructure:"secondary_storage" required:"false"`
	// A pre-built template UUID. This template is used to produce another template. E.g: Ubuntu template.
	BaseTemplateUUID string `mapstructure:"base_template_uuid" required:"false"`
	// A pre-built iso-image is used by the given ISO-image UUID. If IsoImageUUID is set, IsoImageURL is ignored.
	IsoImageUUID string `mapstructure:"isoimage_uuid" required:"false"`
	// An URL is used to download the image. If IsoImageUUID is set, IsoImageURL is ignored.
	IsoImageURL string `mapstructure:"isoimage_url" required:"false"`
	// This is an array of commands to type when the server instance is first
	// booted. The goal of these commands should be to type just enough to
	// initialize the operating system installer. Special keys can be typed as
	// well, and are covered in the section below on the boot command. If this
	// is not specified, it is assumed the installer will start itself.
	BootCommand []string `mapstructure:"boot_command" required:"false"`
	// The time to wait after booting the initial virtual machine before typing
	// the `boot_command`. The value of this should be a duration. Examples are
	// `5s` and `1m30s` which will cause Packer to wait five seconds and one
	// minute 30 seconds, respectively. If this isn't specified, the default is
	// `10s` or 10 seconds. To set boot_wait to 0s, use a negative number, such
	// as "-1s"
	BootWait time.Duration `mapstructure:"boot_wait" required:"false"`
	// Time in ms to wait between each key press
	BootKeyInterval time.Duration `mapstructure:"boot_key_interval" required:"false"`
	// A list of files' relative paths that need to be served on a HTTP server
	Files []string `mapstructure:"files" required:"false"`
	ctx   interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if c.APIURL == "" {
		// Default to environment variable for api_url, if it exists
		c.APIURL = os.Getenv("GRIDSCALE_URL")
	}

	if c.APIToken == "" {
		// Default to environment variable for api_token, if it exists
		c.APIToken = os.Getenv("GRIDSCALE_TOKEN")
	}

	if c.APIKey == "" {
		c.APIKey = os.Getenv("GRIDSCALE_UUID")
	}

	if c.APIRequestHeaders == "" {
		// Default to environment variable for api_request_headers, if it exists
		c.APIRequestHeaders = os.Getenv("GRIDSCALE_PACKER_HEADERS")
	}

	if c.TemplateName == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.TemplateName = def
	}

	if c.ServerName == "" {
		// Default to packer-[time-ordered-uuid]
		c.ServerName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	var errs *packersdk.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}
	if c.APIToken == "" {
		// Required configurations that will display errors if not set
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("api_token for auth must be specified"))
	}
	if c.IsoImageURL == "" && c.IsoImageUUID == "" && c.BaseTemplateUUID == "" {
		errs = packersdk.MultiErrorAppend(
			errs, errors.New("one of these fields has to be set: isoimage_uuid, isoimage_url, base_template_uuid"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packersdk.LogSecretFilter.Set(c.APIToken)
	return c, nil, nil
}
