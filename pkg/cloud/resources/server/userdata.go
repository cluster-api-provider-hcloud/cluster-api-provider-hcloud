package server

import (
	"fmt"

	"github.com/coreos/go-systemd/unit"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var systemdMountTemplate = `[Mount]
What=/dev/disk/by-id/scsi-0HC_Volume_%d
Where=%s
Type=ext4
Options=discard,defaults

[Install]
WantedBy=local-fs.target
`
var systemdUnitAfterRequires = `[Unit]
After=%s
Requires=%s
`

func (s *Service) parseUserData(yamlManifest []byte) (*userData, error) {
	u := &userData{}
	if err := yaml.Unmarshal(yamlManifest, &u.root); err != nil {
		return nil, err
	}
	return u, nil
}

type userDataFile struct {
	Path        string `yaml:"path"`
	Content     string `yaml:"content"`
	Permissions string `yaml:"permissions"`
	Owner       string `yaml:"owner"`
}

type userData struct {
	root yaml.Node
}

func (u *userData) addWaitForMount(unitPath string, mountPath string) error {
	unitName := fmt.Sprintf("%s.mount", unit.UnitNamePathEscape(mountPath))
	file := userDataFile{
		Path:        fmt.Sprintf("/etc/systemd/system/%s", unitPath),
		Content:     fmt.Sprintf(systemdUnitAfterRequires, unitName, unitName),
		Permissions: "0644",
		Owner:       "root:root",
	}
	return u.addWriteFile(file)
}

func (u *userData) appendKubeadmConfig(data string) error {
	kubeadmNode, err := u.findFile("/tmp/kubeadm.yaml")
	if err != nil {
		return err
	}
	for pos := range kubeadmNode.Content {
		if kubeadmNode.Content[pos].Value != "content" {
			continue
		}
		if pos+1 >= len(kubeadmNode.Content) {
			continue
		}
		kubeadmNode.Content[pos+1].Value += data
		return nil
	}
	return errors.New("Could not find content field")
}

func (u *userData) findWriteFiles() (*yaml.Node, error) {
	for pos1 := range u.root.Content {
		for pos2 := range u.root.Content[pos1].Content {
			if u.root.Content[pos1].Content[pos2].Value == "write_files" {
				return u.root.Content[pos1].Content[pos2+1], nil
			}
		}
	}
	return nil, errors.New("no write_files node found")
}

func (u *userData) findFile(path string) (*yaml.Node, error) {
	writeFilesNode, err := u.findWriteFiles()
	if err != nil {
		return nil, err
	}
	for _, file := range writeFilesNode.Content {
		for _, content := range file.Content {
			if content.Value == path {
				return file, nil
			}
		}
	}
	return nil, fmt.Errorf("no file '%s' found", path)
}

func (u *userData) addVolumeMount(id int64, mountPath string) error {
	file := userDataFile{
		Path:        fmt.Sprintf("/etc/systemd/system/%s.mount", unit.UnitNamePathEscape(mountPath)),
		Content:     fmt.Sprintf(systemdMountTemplate, id, mountPath),
		Permissions: "0644",
		Owner:       "root:root",
	}
	return u.addWriteFile(file)
}

func (u *userData) addWriteFile(file userDataFile) error {
	writeFilesNode, err := u.findWriteFiles()

	var fileNode yaml.Node
	fileYAML, err := yaml.Marshal(&file)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(fileYAML, &fileNode); err != nil {
		return err
	}

	writeFilesNode.Content = append(writeFilesNode.Content, fileNode.Content[0])
	return nil
}

func (u *userData) output() ([]byte, error) {
	return yaml.Marshal(&u.root)
}
