package gridscale

import (
	"errors"
	"fmt"
	"os"
	"testing"

	builderT "github.com/hashicorp/packer/helper/builder/testing"
	"github.com/hashicorp/packer/packer"
)

const templateName = "ubuntu-test-acc"

func TestBuilderAcc_basic(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: fmt.Sprintf(testBuilderAccBasic, templateName),
		Check:    testAccCheckAfterTemplateCreation,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("GRIDSCALE_UUID"); v == "" {
		t.Fatal("GRIDSCALE_UUID must be set for acceptance tests")
	}
	if v := os.Getenv("GRIDSCALE_TOKEN"); v == "" {
		t.Fatal("GRIDSCALE_TOKEN must be set for acceptance tests")
	}
}

func testAccCheckAfterTemplateCreation(arts []packer.Artifact) error {
	if len(arts) != 1 {
		return errors.New("number of artifacts should be 1")
	}
	if arts[0].Id() == "" {
		return errors.New("Template UUID not found")
	}
	if arts[0].String() != fmt.Sprintf("A template was created: '%v' (ID: %v) ", templateName, arts[0].Id()) {
		return errors.New("got wrong String()")
	}
	return nil
}

const testBuilderAccBasic = `
{
	"builders": [
    {
	  "type": "gridscale",
      "template_name": "%s",
      "password": "testPassword",
      "hostname": "test-hostname",
      "ssh_username": "root",
      "server_memory": 4,
      "server_cores": 2,
      "storage_capacity": 10,
      "root_template_uuid": "4db64bfc-9fb2-4976-80b5-94ff43b1233a"
    }
  ]
}
`
