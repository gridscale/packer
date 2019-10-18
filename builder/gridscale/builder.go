package gridscale

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gridscale/gsclient-go/v3"
	"github.com/hashicorp/hcl/v2/hcldec"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// The unique id for the builder
const BuilderId = "packer.gridscale"
const (
	defaultAPIURL                    = "https://api.gridscale.io"
	defaultGSCDelayIntervalMilliSecs = 1000
	defaultGSCMaxNumberOfRetries     = 5
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return nil, warnings, errs
	}
	b.config = *c

	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	//client := gsclient.NewClient(gsclient.DefaultConfiguration(b.config.APIKey, b.config.APIToken))
	c := b.config
	apiURL := defaultAPIURL
	if c.APIURL != "" {
		apiURL = c.APIURL
	}
	client := gsclient.NewClient(gsclient.NewConfiguration(
		apiURL,
		c.APIKey,
		c.APIToken,
		os.Getenv("PACKER_LOG") != "",
		true,
		defaultGSCDelayIntervalMilliSecs,
		defaultGSCMaxNumberOfRetries,
	))
	// Add debug HTTP Headers if set
	if c.APIRequestHeaders != "" {
		client.WithHTTPHeaders(convertStrToHeaderMap(c.APIRequestHeaders))
	}

	// Build the steps
	steps := []multistep.Step{
		&stepCreateServer{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("gs_%s.pem", b.config.PackerBuildName),
			client:       client,
			config:       &b.config,
			ui:           ui,
		},
		&stepCreateBootStorage{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepLinkServerBootStorage{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepCreateSecondaryStorage{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepLinkServerSecondaryStorage{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepCreateIPAddr{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepLinkServerIPAddr{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepGetPublicNetwork{
			client: client,
			ui:     ui,
		},
		&stepLinkServerPublicNetwork{
			client: client,
			ui:     ui,
		},
		&stepCreateISOImage{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepLinkServerISOImage{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepStartServer{
			client: client,
			ui:     ui,
		},
		&StepVNCConnect{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&StepExecuteBootCommand{
			config: &b.config,
			ui:     ui,
		},
		&stepCleanupISOImageInstallation{
			relClient: client,
			sClient:   client,
			config:    &b.config,
			ui:        ui,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost(b.config.Comm.SSHHost, "server_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&stepCreateSnapshot{
			client: client,
			config: &b.config,
			ui:     ui,
		},
		&stepCreateTemplate{
			client: client,
			config: &b.config,
			ui:     ui,
		},
	}
	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)
	// Run the steps
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("template_uuid"); !ok {
		log.Println("Failed to find template_uuid in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		TemplateName: b.config.TemplateName,
		TemplateUUID: state.Get("template_uuid").(string),
		Client:       client,
	}

	return artifact, nil
}
