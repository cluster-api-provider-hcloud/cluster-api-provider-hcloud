package manifests

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func (m *Manifests) initializeKubectl() (err error) {
	m.kubectlPath, err = exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("error finding kubectl: %w", err)
	}
	m.log.V(1).Info("kubectl found in path", "path", m.kubectlPath)

	// get version of kubectl
	version, err := m.kubectlCmd(context.Background(), "version", "--client", "--short").Output()
	if err != nil {
		return fmt.Errorf("error executing kubectl version: %w", err)
	}
	m.log.V(1).Info("kubectl version", "version", strings.TrimSpace(string(version)))

	return nil
}

func (m *Manifests) kubectlCmd(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, m.kubectlPath, args...)
}
