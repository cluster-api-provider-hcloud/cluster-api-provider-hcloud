package manifests

import (
	"github.com/go-logr/logr"
)

type Manifests struct {
	log                logr.Logger
	manifestConfigPath string

	kubectlPath string
}

func New(log logr.Logger, manifestConfigPath string) *Manifests {
	return &Manifests{
		log:                log,
		manifestConfigPath: manifestConfigPath,
	}
}

func (m *Manifests) Initialize() error {
	if m.manifestConfigPath == "" {
		return nil
	}
	if err := m.initializeKubectl(); err != nil {
		return err
	}
	if err := m.initializeConfig(); err != nil {
		return err
	}
	return nil
}
