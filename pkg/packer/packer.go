package packer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/hetznercloud/hcloud-go/hcloud"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer/api"
)

const envHcloudToken = "HCLOUD_TOKEN"

type Packer struct {
	log              logr.Logger
	packerConfigPath string

	packerPath string

	buildsLock sync.Mutex
	builds     map[string]*build
}

type build struct {
	*exec.Cmd
	terminated bool
	result     error
	stdout     bytes.Buffer
	stderr     bytes.Buffer
}

func (b *build) Start() error {
	b.Cmd.Stdout = &b.stdout
	b.Cmd.Stderr = &b.stderr
	err := b.Cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		b.result = b.Cmd.Wait()
		b.terminated = true
	}()
	return err
}

func New(log logr.Logger) *Packer {
	return &Packer{
		log:    log,
		builds: make(map[string]*build),
	}
}

func (m *Packer) Initialize(machine *infrav1.HcloudMachine) error {

	if strings.HasPrefix(machine.Spec.ImageName, "http://") || strings.HasPrefix(machine.Spec.ImageName, "https://") {
		splitStrings := strings.Split(machine.Spec.ImageName, "/")
		folderName := strings.TrimSuffix(splitStrings[len(splitStrings)-1], ".tar.gz")
		_, err := os.Stat(fmt.Sprintf("/tmp/%s", folderName))

		if os.IsNotExist(err) {
			r, err := DownloadFile("my-image.tar.gz", machine.Spec.ImageName)
			if err != nil {
				return fmt.Errorf("error while downloading tar.gz file from %s: %s", machine.Spec.ImageName, err)
			}

			err = ExtractTarGz(bytes.NewBuffer(r))
			if err != nil {
				return fmt.Errorf("error while getting and unzipping image from %s: %s", machine.Spec.ImageName, err)
			}

			splitStrings := strings.Split(machine.Spec.ImageName, "/")
			folderName := strings.TrimSuffix(splitStrings[len(splitStrings)-1], ".tar.gz")
			m.packerConfigPath = fmt.Sprintf("/tmp/%s/image.json", folderName)

		}
	} else {
		m.packerConfigPath = fmt.Sprintf("/%s-packer-config/image.json", machine.Spec.ImageName)
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
	version, err := m.packerCmd(context.Background(), "-v").Output()
	if err != nil {
		return fmt.Errorf("error executing packer version: %w", err)
	}
	m.log.V(1).Info("packer version", "version", strings.TrimSpace(string(version)))

	return nil
}

func (m *Packer) initializeConfig() (errr error) {
	cmd := m.packerCmd(context.Background(), "validate", m.packerConfigPath)
	cmd.Env = []string{fmt.Sprintf("%s=xxx", envHcloudToken)}
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

// EnsureImage checks if the API has an image already build and if not, it will
// run packer build to create one
func (m *Packer) EnsureImage(ctx context.Context, log logr.Logger, hc api.HcloudClient, parameters *api.PackerParameters) (*infrav1.HcloudImageID, error) {

	hash := parameters.Hash()
	key := fmt.Sprintf("%s%s", infrav1.NameHcloudProviderPrefix, "template-hash")

	// check if build is currently running
	m.buildsLock.Lock()
	defer m.buildsLock.Unlock()
	if b, ok := m.builds[hash]; ok {
		// build still running
		if !b.terminated {
			log.V(1).Info("packer image build still running", "parameters", parameters)
			return nil, nil
		}

		// check if build has been finished with error
		if err := b.result; err != nil {
			delete(m.builds, hash)
			return nil, fmt.Errorf("%v stdout=%s stderr=%s", err, b.stdout.String(), b.stderr.String())
		}

		// remove build as it had been successful
		log.Info("packer image successfully built", "parameters", parameters)
		delete(m.builds, hash)
	}

	// query for an existing image
	var opts hcloud.ImageListOpts
	opts.LabelSelector = fmt.Sprintf("%s==%s", key, hash)
	images, err := hc.ListImages(ctx, opts)
	if err != nil {
		return nil, err
	}

	var image *hcloud.Image
	for pos := range images {
		i := images[pos]
		if i.Status != hcloud.ImageStatusAvailable {
			continue
		}
		if image == nil || i.Created.After(image.Created) {
			image = i
		}
	}

	// image found, return the latest image
	if image != nil {
		var id = infrav1.HcloudImageID(image.ID)
		return &id, nil
	}

	// schedule build of hcloud image
	b := &build{Cmd: m.packerCmd(context.Background(), "build", m.packerConfigPath)}
	b.Env = append(
		parameters.EnvironmentVariables(),
		fmt.Sprintf("%s=%s", envHcloudToken, hc.Token()),
	)
	log.Info("started building packer image", "parameters", parameters)

	if err := b.Start(); err != nil {
		return nil, err
	}

	m.builds[hash] = b
	return nil, nil
}

func ExtractTarGz(gzipStream io.Reader) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: NewReader failed: %s", err)
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %s", err)
		}
		path := "/tmp/" + header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path, 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err)
			}
			outFile.Close()

		default:
			return fmt.Errorf(
				"ExtractTarGz: unknown type: %s in %s",
				header.Typeflag,
				header.Name)
		}

	}
	return nil
}

func DownloadFile(filepath string, url string) ([]byte, error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to create file in DownloadFile: %s", err)
	}

	return body, err
}
