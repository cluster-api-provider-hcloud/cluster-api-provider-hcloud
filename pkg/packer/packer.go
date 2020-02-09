package packer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

type Packer struct {
	log              logr.Logger
	packerConfigPath string

	packerPath string
}

func New(log logr.Logger, packerConfigPath string) *Packer {
	return &Packer{
		log:              log,
		packerConfigPath: packerConfigPath,
	}
}

func (m *Packer) Initialize() error {
	if m.packerConfigPath == "" {
		return nil
	}
	if err := m.initializePacker(); err != nil {
		return err
	}
	if err := m.initializeConfig(); err != nil {
		return err
	}
	return nil
}

func (m *Packer) initializePacker() (err error) {
	m.packerPath, err = exec.LookPath("packer")
	if err != nil {
		return fmt.Errorf("error finding packer: %w", err)
	}
	m.log.V(1).Info("packer found in path", "path", m.packerPath)

	// get version of packer
	version, err := m.packerCmd(context.Background(), "version").Output()
	if err != nil {
		return fmt.Errorf("error executing packer version: %w", err)
	}
	m.log.V(1).Info("packer version", "version", strings.TrimSpace(string(version)))

	return nil
}

func (m *Packer) initializeConfig() (errr error) {
	cmd := m.packerCmd(context.Background(), "validate", m.packerConfigPath)
	cmd.Env = []string{"HCLOUD_TOKEN=xxx"}
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error validating packer config '%s': %s %w", m.packerConfigPath, string(output), err)
	}
	m.log.V(1).Info("packer config successfully validated", "output", strings.TrimSpace(string(output)))

	return nil
}

func (m *Packer) packerCmd(ctx context.Context, args ...string) *exec.Cmd {
	c := exec.CommandContext(ctx, m.packerPath, args...)
	c.Env = []string{}
	return c
}
