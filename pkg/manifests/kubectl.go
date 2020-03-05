package manifests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	clientcmd "k8s.io/client-go/tools/clientcmd"
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

func (m *Manifests) Apply(ctx context.Context, client clientcmd.ClientConfig, extVar map[string]string) error {
	rawConfig, err := client.RawConfig()
	if err != nil {
		return errors.Wrap(err, "error creating rawConfig")
	}

	kubeconfigBytes, err := clientcmd.Write(rawConfig)
	if err != nil {
		return errors.Wrap(err, "error creating kubeconfig")
	}

	kubeconfigFile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return errors.Wrap(err, "error creating kubeconfig")
	}
	defer os.Remove(kubeconfigFile.Name()) // clean up

	if _, err := kubeconfigFile.Write(kubeconfigBytes); err != nil {
		return errors.Wrap(err, "error writing to kubeconfig")
	}
	if err := kubeconfigFile.Close(); err != nil {
		return errors.Wrap(err, "error closing kubeconfig")
	}

	cmdApply := m.kubectlCmd(ctx, "apply", "-f", "-")
	cmdApply.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile.Name()),
	}

	stdIn, err := cmdApply.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "error creating stdin pipe")
	}
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	cmdApply.Stdout = &stdOut
	cmdApply.Stderr = &stdErr

	if err := cmdApply.Start(); err != nil {
		return errors.Wrap(err, "error starting kubectl")
	}

	if err := evaluateJsonnet(stdIn, m.manifestConfigPath, extVar); err != nil {
		return errors.Wrap(err, "error generating manifests")
	}
	if err := stdIn.Close(); err != nil {
		return errors.Wrap(err, "error closing stdin pipe")
	}

	if err := cmdApply.Wait(); err != nil {
		return errors.Wrap(err, "error kubectl apply failed")
	}

	return nil
}
