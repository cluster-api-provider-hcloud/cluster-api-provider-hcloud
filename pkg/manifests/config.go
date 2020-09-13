package manifests

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	jsonnet "github.com/google/go-jsonnet"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests/parameters"
)

func sampleParameters() *parameters.ManifestParameters {
	hcloudNetwork := intstr.FromString("cluster-dev")
	hcloudToken := "my-token"
	kubeAPIServerDomain := ""
	manifests := []string{"hcloudCSI", "metricsServer"}
	_, podCIDRBlock, err := net.ParseCIDR("192.168.0.0/17")
	if err != nil {
		panic(err)
	}

	return &parameters.ManifestParameters{
		HcloudToken:         &hcloudToken,
		HcloudNetwork:       &hcloudNetwork,
		PodCIDRBlock:        podCIDRBlock,
		Manifests:           manifests,
		KubeAPIServerDomain: &kubeAPIServerDomain,
	}
}

func (m *Manifests) initializeConfig() (err error) {

	if err := evaluateJsonnet(ioutil.Discard, m.manifestConfigPath, sampleParameters().ExtVar()); err != nil {
		return err
	}
	m.log.V(1).Info("manifests config successfully validated", "path", m.manifestConfigPath)

	return nil
}

// dump YAML converts a parsed json object into yaml strings
func dumpYAML(enc *yaml.Encoder, i interface{}) error {
	switch v := i.(type) {
	case map[string]interface{}:
		_, apiVersionExists := v["apiVersion"]
		_, kindExists := v["kind"]
		if kindExists && apiVersionExists {
			if err := enc.Encode(v); err != nil {
				return err
			}
			return nil
		}
		// order amp
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if err := dumpYAML(enc, v[k]); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, e := range v {
			if err := dumpYAML(enc, e); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected type %T", v)
	}
	return nil
}

func evaluateJsonnet(out io.Writer, path string, extVars map[string]string) error {
	vm := jsonnet.MakeVM()
	vm.ErrorFormatter.SetColorFormatter(color.New(color.FgRed).Fprintf)

	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{
			filepath.Dir(path),
		},
	})

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	for k, v := range extVars {
		vm.ExtVar(k, v)
	}

	output, err := vm.EvaluateSnippet(path, string(bytes))
	if err != nil {
		return err
	}

	var object interface{}
	if err := json.Unmarshal([]byte(output), &object); err != nil {
		return err
	}

	enc := yaml.NewEncoder(out)
	if err := dumpYAML(enc, object); err != nil {
		return err
	}

	return nil
}
