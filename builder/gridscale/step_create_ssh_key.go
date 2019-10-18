package gridscale

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/gridscale/gsclient-go/v3"
	"golang.org/x/crypto/ssh"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
	client       gsclient.SSHKeyOperator
	config       *Config
	ui           packer.Ui
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := s.client
	ui := s.ui
	c := s.config
	if c.BaseTemplateUUID == "" {
		return multistep.ActionContinue
	}
	ui.Say("Creating temporary ssh key for server...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	priv_der := x509.MarshalPKCS1PrivateKey(priv)
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	c.Comm.SSHPrivateKey = pem.EncodeToMemory(&priv_blk)
	pub, _ := ssh.NewPublicKey(&priv.PublicKey)

	sshKey, err := client.CreateSshkey(context.Background(), gsclient.SshkeyCreateRequest{
		Name:   fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID()),
		Sshkey: string(bytes.Trim(ssh.MarshalAuthorizedKey(pub), "\n")),
	})
	if err != nil {
		err := fmt.Errorf("Error getting temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("ssh_key_uuid", sshKey.ObjectUUID)
	ui.Say(fmt.Sprintf("a SSH-key (%s) has been created", sshKey.ObjectUUID))
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write(pem.EncodeToMemory(&priv_blk)); err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("Error setting permissions of debug key: %s", err))
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	client := s.client
	ui := s.ui
	c := s.config
	if c.BaseTemplateUUID == "" {
		ui.Say("No SSH key UUID detected.")
		return
	}
	// Destroy the created boot storage
	sshKeyUUID, ok := state.Get("ssh_key_uuid").(string)
	if !ok {
		err := errors.New("cannot convert ssh_key_uuid to string")
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}
	if sshKeyUUID == "" {
		ui.Say("No SSH key UUID detected.")
		return
	}
	ui.Say(fmt.Sprintf("Destroying the bSSH-key (%s)...", sshKeyUUID))
	err := client.DeleteSshkey(context.Background(), sshKeyUUID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error destroying SSH key (%s). Please destroy it manually: %s", sshKeyUUID, err))
		return
	}
	ui.Say(fmt.Sprintf("Destroyed the SSH-key (%s)", sshKeyUUID))
}
