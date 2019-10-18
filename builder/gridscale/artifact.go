package gridscale

import (
	"context"
	"fmt"
	"github.com/gridscale/gsclient-go/v3"
	"log"
)

type Artifact struct {
	// The name of the template
	TemplateName string

	// The UUID of the template
	TemplateUUID string

	// The client for making API calls
	Client gsclient.TemplateOperator
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with DigitalOcean
	return nil
}

func (a *Artifact) Id() string {
	return a.TemplateUUID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A template was created: '%v' (ID: %v) ", a.TemplateName, a.TemplateUUID)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying template: %s (%s)", a.TemplateName, a.TemplateUUID)
	err := a.Client.DeleteTemplate(context.Background(), a.TemplateUUID)
	return err
}
