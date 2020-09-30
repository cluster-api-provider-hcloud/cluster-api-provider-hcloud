package userdata

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	kubejson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/types/v1beta1"
	kubeyaml "sigs.k8s.io/yaml"
)

const kubeadmConfigurationPathInit = "/tmp/kubeadm.yaml"
const kubeadmConfigurationPathJoin = "/tmp/kubeadm-join-config.yaml"

type UserData struct {
	document yaml.Node
}

type KubeadmConfig struct {
	ClusterConfiguration *kubeadmv1beta1.ClusterConfiguration
	InitConfiguration    *kubeadmv1beta1.InitConfiguration
	JoinConfiguration    *kubeadmv1beta1.JoinConfiguration
	isInit               bool
}

func (k *KubeadmConfig) IsInit() bool {
	return k.isInit
}

func (k *KubeadmConfig) IsJoin() bool {
	return !k.isInit
}

func newScheme() *runtime.Scheme {
	sch := runtime.NewScheme()
	sch.AddKnownTypes(
		kubeadmv1beta1.GroupVersion,
		&kubeadmv1beta1.ClusterConfiguration{},
		&kubeadmv1beta1.InitConfiguration{},
		&kubeadmv1beta1.JoinConfiguration{},
	)
	return sch
}

func newSerializer() *kubejson.Serializer {
	sch := newScheme()
	return kubejson.NewSerializerWithOptions(
		kubejson.DefaultMetaFactory,
		sch,
		sch,
		kubejson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
}

func NewFromReader(r io.Reader) (*UserData, error) {
	d := yaml.NewDecoder(r)
	u := &UserData{}

	if err := d.Decode(&u.document); err != nil {
		return nil, fmt.Errorf("error parsing YAML user-data: %w", err)
	}

	return u, nil
}

func (u *UserData) WriteYAML(w io.Writer) error {
	return yaml.NewEncoder(w).Encode(&u.document)
}

func nextNode(nodes []*yaml.Node, pos int) *yaml.Node {
	if len(nodes) > pos+1 {
		return nodes[pos+1]
	}
	return nil
}

func (u *UserData) getWriteFiles() (*yaml.Node, error) {
	for _, l1 := range u.document.Content {
		for pos, l2 := range l1.Content {
			if next := nextNode(l1.Content, pos); l2.Value == "write_files" && next != nil {
				return next, nil
			}
		}
	}

	return nil, fmt.Errorf("write_files not found in cloud-init")
}

func (u *UserData) getFile(path string) (*yaml.Node, error) {
	// find write files field
	writeFiles, err := u.getWriteFiles()
	if err != nil {
		return nil, err
	}

	// find the correct path
	for _, l1 := range writeFiles.Content {
		for pos, l2 := range l1.Content {
			if next := nextNode(l1.Content, pos); l2.Value == "path" && next != nil && next.Value == path {
				return l1, nil
			}
		}
	}

	// file not found
	return nil, nil

}

func parseKubeadmConfig(data []byte) (*KubeadmConfig, error) {
	//var typeMeta metav1.TypeMeta
	k := &KubeadmConfig{}

	serializer := newSerializer()

	reader := kubejson.YAMLFramer.NewFrameReader(ioutil.NopCloser(bytes.NewReader([]byte(data))))
	d := streaming.NewDecoder(reader, serializer)

	for {
		obj, _, err := d.Decode(nil, nil)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error during parse: %w", err)
		}
		switch o := obj.(type) {
		case *kubeadmv1beta1.ClusterConfiguration:
			k.ClusterConfiguration = o
		case *kubeadmv1beta1.InitConfiguration:
			k.InitConfiguration = o
		case *kubeadmv1beta1.JoinConfiguration:
			k.JoinConfiguration = o
		default:
			return nil, fmt.Errorf("unknown type during parse: %v", o)
		}
	}

	return k, nil
}

func (u *UserData) GetKubeadmConfig() (*KubeadmConfig, error) {
	isInit := true

	n, err := u.getFile(kubeadmConfigurationPathInit)
	if err != nil {
		return nil, err
	}
	if n == nil {
		n, err = u.getFile(kubeadmConfigurationPathJoin)
		if err != nil {
			return nil, err
		}
		if n != nil {
			isInit = false
		}
	}

	kubeadmConfig := &KubeadmConfig{}

	if n != nil {
		for pos, l1 := range n.Content {
			if next := nextNode(n.Content, pos); l1.Value == "content" && next != nil {
				k, err := parseKubeadmConfig([]byte(next.Value))
				if err != nil {
					return nil, err
				}
				kubeadmConfig = k
			}
		}
	}
	kubeadmConfig.isInit = isInit
	return kubeadmConfig, nil
}

func (u *UserData) SetKubeadmConfig(k *KubeadmConfig) error {
	codecs := serializer.NewCodecFactory(newScheme())

	mediaType := "application/yaml"
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	var objects []string

	addObj := func(s string) {
		objects = append(objects, "---\n"+s)
	}

	kubeadmEncoder := codecs.EncoderForVersion(info.Serializer, kubeadmv1beta1.GroupVersion)

	if obj := k.ClusterConfiguration; obj != nil {
		b := bytes.NewBuffer(nil)
		if err := kubeadmEncoder.Encode(obj, b); err != nil {
			return fmt.Errorf("error serializing %T: %w", obj, err)
		}
		addObj(b.String())
	}

	if obj := k.InitConfiguration; obj != nil {
		b := bytes.NewBuffer(nil)
		if err := kubeadmEncoder.Encode(obj, b); err != nil {
			return fmt.Errorf("error serializing %T: %w", obj, err)
		}
		addObj(b.String())
	}

	if obj := k.JoinConfiguration; obj != nil {
		b := bytes.NewBuffer(nil)
		if err := kubeadmEncoder.Encode(obj, b); err != nil {
			return fmt.Errorf("error serializing %T: %w", obj, err)
		}
		addObj(b.String())
	}

	path := kubeadmConfigurationPathInit
	if k.IsJoin() {
		path = kubeadmConfigurationPathJoin
	}

	return u.SetOrUpdateFile(bootstrapv1.File{
		Path:    path,
		Content: strings.Join(objects, ""),
	})
}

func (u *UserData) SetOrUpdateFile(file bootstrapv1.File) error {
	n, err := u.getFile(file.Path)
	if err != nil {
		return err
	}

	if n == nil {
		// encode to json and then parse using YAMLv3
		fileYAML, err := kubeyaml.Marshal(&file)
		if err != nil {
			return err
		}

		var fileNode yaml.Node
		dec := yaml.NewDecoder(bytes.NewReader(fileYAML))
		if err := dec.Decode(&fileNode); err != nil {
			return err
		}

		// find write files field
		writeFiles, err := u.getWriteFiles()
		if err != nil {
			return err
		}

		writeFiles.Content = append(writeFiles.Content, fileNode.Content[0])
		return nil
	}

	updateFields := make(map[string]string)
	if len(file.Content) > 0 {
		updateFields["content"] = file.Content
	}
	if len(file.Permissions) > 0 {
		updateFields["permissions"] = file.Permissions
	}
	if len(file.Owner) > 0 {
		updateFields["owner"] = file.Owner
	}

	for pos, l1 := range n.Content {
		if pos%2 == 1 {
			continue
		}
		value, ok := updateFields[l1.Value]
		if !ok {
			continue
		}

		if next := nextNode(n.Content, pos); next != nil {
			next.Value = value
			delete(updateFields, l1.Value)
		}
	}

	if len(updateFields) == 0 {
		return nil
	}

	var keys []string
	for k := range updateFields {
		keys = append(keys, k)
	}

	return fmt.Errorf("existing file did not have %v keys", strings.Join(keys, ", "))
}

func (u *UserData) SkipKubeProxy() error {
	var n *yaml.Node
	for _, l1 := range u.document.Content {
		for pos, l2 := range l1.Content {
			if next := nextNode(l1.Content, pos); l2.Value == "runcmd" && next != nil {
				n = next
			}
		}
	}
	if n == nil {
		return errors.New("runcmd variable not found")
	}
	for _, l1 := range n.Content {
		if strings.HasPrefix(l1.Value, "kubeadm init") {
			flag := "--skip-phases=addon/kube-proxy"
			if strings.Contains(l1.Value, flag) {
				return nil
			}
			l1.Value = fmt.Sprintf("%s %s", strings.TrimRight(l1.Value, " "), flag)
			return nil
		}
	}
	return errors.New("kubeadm init command not found")
}

func (u *UserData) GetContentInformation() (string, error) {

	if len(u.document.Content) > 0 &&
		len(u.document.Content[0].Content) > 1 &&
		len(u.document.Content[0].Content[1].Content) > 0 &&
		len(u.document.Content[0].Content[1].Content[0].Content) > 6 {
		return u.document.Content[0].Content[1].Content[0].Content[7].Value, nil
	}
	return "", errors.New("Content of yaml node does not exist")
}
