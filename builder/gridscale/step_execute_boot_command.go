package gridscale

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/go-vnc"
)

const defaultBootWaitSecs = 120

type StepExecuteBootCommand struct {
	config *Config
	ui     packer.Ui
}

func (s StepExecuteBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := s.ui
	c := s.config
	if len(c.BootCommand) == 0 {
		ui.Say("boot_command is not set. Skipping executing VNC boot commands...")
		return multistep.ActionContinue
	}
	conn := state.Get("vnc_conn").(*vnc.ClientConn)
	defer conn.Close()

	// Wait the for the vm to boot.
	bootWait := defaultBootWaitSecs * time.Second
	if int64(c.BootWait) > 0 {
		bootWait = c.BootWait
	}
	ui.Say(fmt.Sprintf("Waiting %s for boot...", bootWait.String()))
	select {
	case <-time.After(bootWait):
		break
	case <-ctx.Done():
		return multistep.ActionHalt
	}
	d := bootcommand.NewVNCDriver(conn, c.BootKeyInterval)

	ui.Say("Typing the boot command over VNC...")
	flatBootCommand := strings.Join(c.BootCommand, "")
	command, err := interpolate.Render(flatBootCommand, &c.ctx)
	if err != nil {
		err := fmt.Errorf("Error preparing boot command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	seq, err := bootcommand.GenerateExpressionSequence(command)
	if err != nil {
		err := fmt.Errorf("Error generating boot command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := seq.Do(ctx, d); err != nil {
		err := fmt.Errorf("Error running boot command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Finished executing boot command")
	return multistep.ActionContinue
}

func (s StepExecuteBootCommand) Cleanup(state multistep.StateBag) {
}
