package baremetal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/userdata"
	"github.com/nl2go/hrobot-go/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	capierrors "sigs.k8s.io/cluster-api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	delimiter                      string        = "--"
	hoursBeforeDeletion            time.Duration = 36
	maxWaitingTimeForSSHConnection int           = 200
	defaultPort                    int           = 22
	rateLimitTimeOut               time.Duration = 660
	rateLimitTimeOutDeletion       time.Duration = 120
)

type Service struct {
	scope *scope.HetznerBareMetalMachineScope
}

type blockDeviceData struct {
	Name     string `json:"name"`
	Size     string `json:"size"`
	Rotation bool   `json:"rota"`
	Type     string `json:"fstype"`
	Label    string `json:"label"`
}

type blockDevice struct {
	Name     string            `json:"name"`
	Size     string            `json:"size"`
	Rotation bool              `json:"rota"`
	Type     string            `json:"fstype"`
	Label    string            `json:"label"`
	Children []blockDeviceData `json:"children"`
}

type blockDevices struct {
	Devices []blockDevice `json:"blockdevices"`
}

func NewService(scope *scope.HetznerBareMetalMachineScope) *Service {
	return &Service{
		scope: scope,
	}
}

func stringSliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {

	// If not token information has been given, the server cannot be successfully reconciled
	if s.scope.HcloudCluster.Spec.HrobotTokenRef == nil {
		s.scope.Recorder.Eventf(s.scope.HetznerBareMetalMachine, corev1.EventTypeWarning, "NoTokenFound", "No Hrobot token found")
		return nil, errors.Errorf("ERROR: No token for Hetzner Robot provided: Cannot reconcile server %s", s.scope.HetznerBareMetalMachine.Name)
	}

	// update list of servers which have the right type and are not taken in other clusters
	serverList, err := s.listMatchingMachines(ctx)
	if err != nil {
		if checkRateLimitExceeded(err) {
			s.scope.Recorder.Eventf(s.scope.HetznerBareMetalMachine, corev1.EventTypeWarning, "HrobotRateLimitExceeded", "Hrobot rate limit exceeded. Wait for %v sec before trying again.", rateLimitTimeOut)
			return &reconcile.Result{RequeueAfter: rateLimitTimeOut * time.Second}, nil
		}
		return nil, errors.Wrap(err, "failed to list machines")
	}

	// Find the machine if already attached to the cluster
	actualServer, err := s.findAttachedMachine(serverList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find attached machine")
	}

	// There exists an attached machine
	if actualServer != nil {
		// check if server has been cancelled
		if actualServer.Cancelled {
			s.scope.V(1).Info("server has been cancelled", "server", actualServer.ServerName, "cancelled",
				actualServer.Cancelled, "paid until", actualServer.PaidUntil)
			paidUntil, err := time.Parse("2006-01-02", actualServer.PaidUntil)
			if err != nil {
				return nil, errors.Errorf("ERROR: Failed to parse paidUntil date. Error: %s", err)
			}
			// If the server will be cancelled soon but is still in the list of actualServers, it means that
			// it is already attached since we don't add new servers that get deleted soon to actualServers
			// This means we have to set a failureReason so that the infrastructure object gets deleted
			if paidUntil.Before(time.Now().Add(hoursBeforeDeletion * time.Hour)) {
				s.scope.Recorder.Eventf(
					s.scope.HetznerBareMetalMachine,
					corev1.EventTypeWarning,
					"CancelledHetznerBareMetalMachine",
					"Bare metal machine is cancelled and paid until %s",
					actualServer.PaidUntil)
				er := capierrors.UpdateMachineError
				s.scope.HetznerBareMetalMachine.Status.FailureReason = &er
				s.scope.HetznerBareMetalMachine.Status.FailureMessage = pointer.StringPtr("Machine has been cancelled and is paid until less than" + hoursBeforeDeletion.String() + "hours")
				s.scope.HetznerBareMetalMachine.Status.Ready = false
			}
		}

	} else {
		// If no machine with the correct name is already attached, we have to find one which suits our needs,
		// i.e. which is not attached already
		newServer, err := s.findNewMachine(serverList)
		if err != nil {
			return nil, errors.Errorf("Failed to find new machine: %s", err)
		}
		// If the machine is not ready to provision yet, we wait
		if newServer.Status != "ready" {
			s.scope.V(2).Info("server not in running state", "server", newServer.ServerName, "status", newServer.Status)
			return &reconcile.Result{RequeueAfter: 300 * time.Second}, nil
		}
		// Wait for Bootstrap data to be ready
		if !s.scope.IsBootstrapDataReady(ctx) {
			s.scope.V(2).Info("Bootstrap data is not ready yet")
			return &reconcile.Result{RequeueAfter: 15 * time.Second}, nil
		}

		s.scope.HetznerBareMetalMachine.Status.ServerState = "initializing"

		// Provision the machine
		if err := s.provisionMachine(ctx, *newServer); err != nil {
			if checkRateLimitExceeded(err) {
				s.scope.Recorder.Eventf(s.scope.HetznerBareMetalMachine, corev1.EventTypeWarning, "HrobotRateLimitExceeded", "Hrobot rate limit exceeded. Wait for %v sec before trying again.", rateLimitTimeOut)
				return &reconcile.Result{RequeueAfter: rateLimitTimeOut * time.Second}, nil
			}
			return nil, errors.Errorf("Failed to provision new machine: %s", err)
		}
		actualServer = newServer
	}

	providerID := fmt.Sprintf("hcloud://%d", actualServer.ServerNumber)
	s.scope.HetznerBareMetalMachine.Status.ServerState = "running"
	s.scope.HetznerBareMetalMachine.Spec.ProviderID = &providerID

	// TODO: Ask for the state of the server and only if it is ready set it to true
	s.scope.HetznerBareMetalMachine.Status.Ready = true

	return nil, nil
}

// looks if a machine of the correct name has been attached already
func (s *Service) findAttachedMachine(servers []models.Server) (*models.Server, error) {

	var actualServer models.Server

	// If the list of servers is empty, then we cannot find an attached machine
	if len(servers) == 0 {
		return nil, nil
	}

	var check int
	for _, server := range servers {
		splitName := strings.Split(server.ServerName, delimiter)

		// If server is attached to the cluster and is the one we are looking for
		if len(splitName) == 3 && splitName[2] == s.scope.HetznerBareMetalMachine.Name {
			actualServer = server
			check++
		}
	}
	if check > 1 {
		s.scope.Recorder.Eventf(
			s.scope.HetznerBareMetalMachine,
			corev1.EventTypeWarning,
			"MultipleHetznerBareMetalMachines",
			"Found %v bare metal machines of the name %s attached to the cluster",
			check, actualServer.ServerName)
		return nil, errors.Errorf("There are %s servers which are attached to the cluster with name %s", check, actualServer.ServerName)
	} else if check == 0 {
		// No attached server with the correct name found
		return nil, nil
	} else {
		return &actualServer, nil
	}
}

func (s *Service) findNewMachine(servers []models.Server) (*models.Server, error) {

	if len(servers) == 0 {
		// If no servers are in the list, we have to return an error
		s.scope.Recorder.Eventf(
			s.scope.HetznerBareMetalMachine,
			corev1.EventTypeWarning,
			"EmptyListHetznerBareMetalMachines",
			"No machines of type %s found",
			*s.scope.HetznerBareMetalMachine.Spec.ServerType)
		return nil, errors.Errorf("No bare metal server found with type %s", *s.scope.HetznerBareMetalMachine.Spec.ServerType)
	}

	for _, server := range servers {
		splitName := strings.Split(server.ServerName, delimiter)
		// If the name gets split into two parts in our structure cluster_prefix -- type -- name
		// it means that the server is not attached to any cluster yet
		if len(splitName) == 2 {
			return &server, nil
		}
	}
	s.scope.Recorder.Eventf(
		s.scope.HetznerBareMetalMachine,
		corev1.EventTypeWarning,
		"NoAvailableHetznerBareMetalMachines",
		"No machines of type %s are available",
		*s.scope.HetznerBareMetalMachine.Spec.ServerType)
	return nil, errors.Errorf("There are no available servers of type %s left", *s.scope.HetznerBareMetalMachine.Spec.ServerType)
}

// Provisions the bare metal machine
func (s *Service) provisionMachine(ctx context.Context, server models.Server) error {
	s.scope.V(4).Info("Started provisioning bare metal machine")

	// We use SSH so the keys must be specified in a secret
	sshKeyName, _, privateSSHKey, err := s.retrieveSSHSecret(ctx)
	if err != nil {
		return errors.Errorf("Unable to retrieve SSH secret: ", err)
	}

	sshFingerprint, err := s.getSSHFingerprintFromName(sshKeyName)
	if err != nil {
		return errors.Errorf("Unable to get SSH fingerprint for the SSH key %s: %s ", sshKeyName, err)
	}

	// Get the Kubeadm config from the bootstrap provider
	userDataInitial, err := s.scope.GetRawBootstrapData(ctx)
	if err != nil {
		return err
	}
	userData, err := userdata.NewFromReader(bytes.NewReader(userDataInitial))
	if err != nil {
		return err
	}

	kubeadmConfig, err := userData.GetKubeadmConfig()
	if err != nil {
		return err
	}

	cloudProviderKey := "cloud-provider"
	cloudProviderValue := "external"

	if j := kubeadmConfig.JoinConfiguration; j != nil {
		if j.NodeRegistration.KubeletExtraArgs == nil {
			j.NodeRegistration.KubeletExtraArgs = make(map[string]string)
		}
		if _, ok := j.NodeRegistration.KubeletExtraArgs[cloudProviderKey]; !ok {
			j.NodeRegistration.KubeletExtraArgs[cloudProviderKey] = cloudProviderValue
		}
	}

	if err := userData.SetKubeadmConfig(kubeadmConfig); err != nil {
		return err
	}

	userDataBytes := bytes.NewBuffer(nil)
	if err := userData.WriteYAML(userDataBytes); err != nil {
		return errors.Errorf("Error while writing yaml file", err)
	}

	cloudInitConfigString := userDataBytes.String()
	s.scope.V(4).Info("Activate rescue")
	// First we have to activate rescue mode
	_, err = s.scope.HrobotClient().ActivateRescue(server.ServerIP, sshFingerprint)
	if err != nil {
		return errors.Errorf("Unable to activate rescue system: ", err)
	}
	s.scope.V(4).Info("Reset machine")
	// reboot system
	_, err = s.scope.HrobotClient().ResetBMServer(server.ServerIP, "hw")
	s.scope.V(4).Info("Run SSH command hostname")
	// Find out if rescue system has been started successfully
	stdout, stderr, err := runSSH("hostname", server.ServerIP, 22, privateSSHKey)
	if err != nil {
		return errors.Errorf("SSH command hostname returned the error %s. The output of stderr is %s", err, stderr)
	}
	if !strings.Contains(stdout, "rescue") {
		return errors.Errorf("Rescue system not successfully started. Output of command hostname is %s", stdout)
	}

	// Find the block device where we can install our image

	s.scope.V(4).Info("Get block devices")

	var blockDevices blockDevices
	blockDeviceCommand := "lsblk -o name,size,rota,fstype,label -e1 -e7 --json"
	stdout, stderr, err = runSSH(blockDeviceCommand, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Error running the ssh command %s: Error: %s, stderr: %s", blockDeviceCommand, err, stderr)
	}

	err = json.Unmarshal([]byte(stdout), &blockDevices)
	if err != nil {
		return errors.Errorf("Error while unmarshaling: %s", err)
	}

	drive, err := findCorrectDevice(blockDevices)
	if err != nil {
		return errors.Errorf("Error while finding correct device: %s", err)
	}

	partitionString := `PART /boot ext3 512M
PART / ext4 all`
	if s.scope.HetznerBareMetalMachine.Spec.Partition != nil {
		partitionString = *s.scope.HetznerBareMetalMachine.Spec.Partition
	}
	autoSetup := fmt.Sprintf(
		`cat > /autosetup << EOF
DRIVE1 /dev/%s
BOOTLOADER grub
HOSTNAME %s
%s
IMAGE %s
EOF`,
		drive, s.scope.Cluster.Name+delimiter+*s.scope.HetznerBareMetalMachine.Spec.ServerType+delimiter+s.scope.HetznerBareMetalMachine.Name, partitionString, *s.scope.HetznerBareMetalMachine.Spec.ImagePath)

	s.scope.V(4).Info("Send auto setup file to server")
	// Send autosetup file to server
	_, stderr, err = runSSH(autoSetup, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("SSH command autosetup returned the error %s. The output of stderr is %s", err, stderr)
	}

	s.scope.V(4).Info("Install image")
	// Install the image
	_, stderr, err = runSSH("bash /root/.oldroot/nfs/install/installimage", server.ServerIP, 22, privateSSHKey)
	if err != nil {
		// If an error occurs here, we have to wipe the device to avoid future problems
		wipeCommand := fmt.Sprintf("wipefs -a /dev/%s", drive)
		_, _, _ = runSSH(wipeCommand, server.ServerIP, defaultPort, privateSSHKey)
		return errors.Errorf("SSH command installimage returned the error %s. The output of stderr is %s", err, stderr)
	}
	s.scope.V(4).Info("Get block devices")
	// get again list of block devices and label children of our drive
	stdout, stderr, err = runSSH(blockDeviceCommand, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		// If an error occurs here, we have to wipe the device to avoid future problems
		wipeCommand := fmt.Sprintf("wipefs -a /dev/%s", drive)
		_, _, _ = runSSH(wipeCommand, server.ServerIP, defaultPort, privateSSHKey)
		return errors.Errorf("Error running the ssh command %s: Error: %s, stderr: %s", blockDeviceCommand, err, stderr)
	}

	err = json.Unmarshal([]byte(stdout), &blockDevices)
	if err != nil {
		return errors.Errorf("Error while unmarshaling: %s", err)
	}

	// label the block device so that we know in the future where we installed our image
	command, err := labelChildrenCommand(blockDevices, drive)
	if err != nil {
		return errors.Errorf("Error while constructing labeling children command of device %s: %s", drive, err)
	}
	s.scope.V(4).Info("Label children")
	_, stderr, err = runSSH(command, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		// If an error occurs here, we have to wipe the device to avoid future problems
		wipeCommand := fmt.Sprintf("wipefs -a /dev/%s", drive)
		_, _, _ = runSSH(wipeCommand, server.ServerIP, defaultPort, privateSSHKey)
		return errors.Errorf("Error running the ssh command %s: Error: %s, stderr: %s", command, err, stderr)
	}

	// avoid errors when reboot comes too early for the previous command
	_, stderr, err = runSSH("sleep 30", server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Error running the ssh command sleep 30: Error: %s, stderr: %s", err, stderr)
	}

	s.scope.V(4).Info("Reboot")
	// reboot system
	_, stderr, err = runSSH("reboot", server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		if !strings.Contains(err.Error(), "exited without exit status or exit signal") {
			return errors.Errorf("Error running the ssh command reboot: Error: %s, stderr: %s", err, stderr)
		}
	}

	// We cannot create the files right after rebooting, as it gets deleted again
	// so we have to wait for a bit
	_, stderr, err = runSSH("sleep 60", server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Connection to server after reboot could not be established. Error: %s, stderr: %s", err, stderr)
	}
	s.scope.V(4).Info("Create nocloud directory")
	// create nocloud directory for cloud-init
	command = "mkdir -p /var/lib/cloud/seed/nocloud"
	_, stderr, err = runSSH(command, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Error running the ssh command %s: Error: %s, stderr: %s", command, err, stderr)
	}

	s.scope.V(4).Info("Create meta data for cloud init")
	// create meta-data for cloud-init
	command = "echo 'instance-id: iid-system-uuid' >> /var/lib/cloud/seed/nocloud/meta-data"
	_, stderr, err = runSSH(command, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Error running the ssh command %s: Error: %s, stderr: %s", command, err, stderr)
	}

	// create user-data for cloud-init provider nocloud
	cloudInitCommand := fmt.Sprintf(
		`cat > /var/lib/cloud/seed/nocloud/user-data << EOF
%s
EOF`, cloudInitConfigString)

	s.scope.V(4).Info("Send userdata")
	// send user-data to server
	_, stderr, err = runSSH(cloudInitCommand, server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Error running the ssh command %s: Error: %s, stderr: %s", cloudInitCommand, err, stderr)
	}

	// Wait for server
	_, stderr, err = runSSH("sleep 30", server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		return errors.Errorf("Connection to server after reboot could not be established. Error: %s, stderr: %s", err, stderr)
	}

	s.scope.V(4).Info("Reboot")
	// reboot system
	_, stderr, err = runSSH("reboot", server.ServerIP, defaultPort, privateSSHKey)
	if err != nil {
		if !strings.Contains(err.Error(), "exited without exit status or exit signal") {
			return errors.Errorf("Error running the ssh command reboot: Error: %s, stderr: %s", err, stderr)
		}
	}
	s.scope.V(4).Info("Set server name and finish")
	// Finally set the machine's name. The name replaces labels as we cannot label bare metal machines directly
	_, err = s.scope.HrobotClient().SetBMServerName(server.ServerIP,
		s.scope.Cluster.Name+delimiter+*s.scope.HetznerBareMetalMachine.Spec.ServerType+delimiter+s.scope.HetznerBareMetalMachine.Name)
	if err != nil {
		return errors.Errorf("Unable to change bare metal server name: ", err)
	}
	s.scope.Recorder.Eventf(
		s.scope.HetznerBareMetalMachine,
		"CreateHetznerBareMetalMachine",
		"Created bare metal machine %s",
		s.scope.HetznerBareMetalMachine.Name)
	return nil
}

func checkRateLimitExceeded(err error) bool {
	return strings.Contains(err.Error(), "rate limit exceeded")
}

// In order to avoid the loss of data, all we do is renaming the machine in Hrobot.
// After re-attaching the machine we know on which device we installed the image
func (s *Service) Delete(ctx context.Context) (_ *ctrl.Result, err error) {

	serverList, err := s.listMatchingMachines(ctx)
	if err != nil {
		if checkRateLimitExceeded(err) {
			s.scope.Recorder.Eventf(s.scope.HetznerBareMetalMachine, corev1.EventTypeWarning, "HrobotRateLimitExceeded", "Hrobot rate limit exceeded. Wait for %v sec before trying again.", rateLimitTimeOutDeletion)
			return &reconcile.Result{RequeueAfter: rateLimitTimeOutDeletion * time.Second}, nil
		}
		return nil, errors.Wrap(err, "failed to refresh server status")
	}

	server, err := s.findAttachedMachine(serverList)
	if err != nil {
		if checkRateLimitExceeded(err) {
			s.scope.Recorder.Eventf(s.scope.HetznerBareMetalMachine, corev1.EventTypeWarning, "HrobotRateLimitExceeded", "Hrobot rate limit exceeded. Wait for %v sec before trying again.", rateLimitTimeOutDeletion)
			return &reconcile.Result{RequeueAfter: rateLimitTimeOutDeletion * time.Second}, nil
		}
		return nil, errors.Wrap(err, "failed to find attached machine")
	}
	if server == nil {
		s.scope.Recorder.Eventf(
			s.scope.HetznerBareMetalMachine,
			"UnknownHetznerBareMetalMachine",
			"No machine with name %s found to delete",
			s.scope.HetznerBareMetalMachine.Name)
		return nil, nil
	}

	_, err = s.scope.HrobotClient().SetBMServerName(server.ServerIP,
		*s.scope.HetznerBareMetalMachine.Spec.ServerType+delimiter+"unused-"+s.scope.HetznerBareMetalMachine.Name)
	if err != nil {
		return nil, errors.Errorf("Unable to change bare metal server name: ", err)
	}

	return nil, nil
}

func (s *Service) getSSHFingerprintFromName(name string) (fingerprint string, err error) {

	sshKeys, err := s.scope.HrobotClient().ListBMKeys()
	if len(sshKeys) == 0 {
		return "", errors.New("No SSH Keys given")
	}

	var found bool
	for _, key := range sshKeys {
		if name == key.Name {
			fingerprint = key.Fingerprint
			found = true
		}
	}

	if found == false {
		return "", errors.Errorf("No SSH key with name %s found", name)
	}

	return fingerprint, nil
}

// listMatchingMachines looks for all machines that have the right type and which are
// either attached to this cluster or not attached to any cluster yet
func (s *Service) listMatchingMachines(ctx context.Context) ([]models.Server, error) {

	serverList, err := s.scope.HrobotClient().ListBMServers()
	if err != nil {
		s.scope.Recorder.Eventf(
			s.scope.HetznerBareMetalMachine,
			corev1.EventTypeWarning,
			"ErrorListingHetznerBareMetalMachines",
			"Error while listing bare metal machines of type %s",
			*s.scope.HetznerBareMetalMachine.Spec.ServerType)
		return nil, errors.Errorf("unable to list bare metal servers: %s", err)
	}

	var matchingServers []models.Server
	// We exclude all machines that are not yet attached and only paid until very soon
	// as well as machines that don't fit the type or are attached to other clusters
	for _, server := range serverList {
		splitName := strings.Split(server.ServerName, delimiter)
		if len(splitName) == 2 && splitName[0] == *s.scope.HetznerBareMetalMachine.Spec.ServerType {
			if !server.Cancelled {
				matchingServers = append(matchingServers, server)
			} else {
				paidUntil, err := time.Parse("2006-01-02", server.PaidUntil)
				if err != nil {
					return nil, errors.Errorf("ERROR: Failed to parse paidUntil date. Error: %s", err)
				}
				if !paidUntil.Before(time.Now().Add(hoursBeforeDeletion * time.Hour)) {
					matchingServers = append(matchingServers, server)
				}
			}
		}
		if len(splitName) == 3 && splitName[0] == s.scope.Cluster.Name && splitName[1] == *s.scope.HetznerBareMetalMachine.Spec.ServerType {
			matchingServers = append(matchingServers, server)
		}
	}

	return matchingServers, nil
}

func labelChildrenCommand(blockDevices blockDevices, drive string) (string, error) {

	var device blockDevice
	var check bool
	for _, d := range blockDevices.Devices {
		if drive == d.Name {
			device = d
			check = true
		}
	}

	if check == false {
		return "", errors.Errorf("no device with name %s found", drive)
	}

	if device.Children == nil {
		return "", errors.Errorf("no children for device with name %s found, instalimage did not work properly", drive)
	}

	var command string
	for _, child := range device.Children {
		str := fmt.Sprintf(`e2label /dev/%s os
`, child.Name)
		command = command + str
	}
	command = command[:len(command)-1]
	return command, nil
}

func findCorrectDevice(blockDevices blockDevices) (drive string, err error) {
	// If no blockdevice has correctly labeled children, we follow a certain set of rules to find the right one
	var numLabels int
	var hasChildren int
	for _, device := range blockDevices.Devices {
		if device.Children != nil && strings.HasPrefix(device.Name, "sd") {
			hasChildren++
			for _, child := range device.Children {
				if child.Label == "os" {
					numLabels++
					drive = device.Name
					break
				}
			}
		}
	}

	// if numLabels == 1 then finished (drive has been set already)
	// if numLabels == 0 then start sorting
	// if numLabels > 1 then throw error
	var filteredDevices []blockDevice

	if numLabels > 1 {
		return "", errors.Errorf("Found %v devices with the correct labels", numLabels)
	} else if numLabels == 0 {
		// If every device has children, then there is none left for us
		if hasChildren == len(blockDevices.Devices) {
			return "", errors.New("No device is left for installing the operating system")
		}

		// Choose devices with no children, whose name starts with "sd" and which are SSDs (i.e. rota=false)
		for _, device := range blockDevices.Devices {
			if device.Children == nil && strings.HasPrefix(device.Name, "sd") && device.Rotation == false {
				filteredDevices = append(filteredDevices, device)
			}
		}

		// This means that there is no SSD available. Then we have to include HDD as well
		if len(filteredDevices) == 0 {
			for _, device := range blockDevices.Devices {
				if device.Children == nil && strings.HasPrefix(device.Name, "sd") {
					filteredDevices = append(filteredDevices, device)
				}
			}
			// If there is only one device which satisfies the requirements then we choose it
		} else if len(filteredDevices) == 1 {
			drive = filteredDevices[0].Name
			// If there are more devices then we need to sort them according to our specifications
		} else {
			// First change the data type of size, so that we can compare it
			type reducedBlockDevice struct {
				Name string
				Size int
			}
			var reducedDevices []reducedBlockDevice
			for _, device := range filteredDevices {
				size, err := convertSizeToInt(device.Size)
				if err != nil {
					return "", errors.Errorf("Could not convert size %s to integer", device.Size)
				}
				reducedDevices = append(reducedDevices, reducedBlockDevice{
					Name: device.Name,
					Size: size,
				})
			}
			// Sort the devices with respect to size
			sort.SliceStable(reducedDevices, func(i, j int) bool {
				return reducedDevices[i].Size < reducedDevices[j].Size
			})

			// Look whether there is more than one device with the same size
			var filteredReducedDevices []reducedBlockDevice
			if reducedDevices[0].Size < reducedDevices[1].Size {
				drive = reducedDevices[0].Name
			} else {
				for _, device := range reducedDevices {
					if device.Size == reducedDevices[0].Size {
						filteredReducedDevices = append(filteredReducedDevices, device)
					}
				}
				// Sort the devices with respect to name
				sort.SliceStable(filteredReducedDevices, func(i, j int) bool {
					return filteredReducedDevices[i].Name > filteredReducedDevices[j].Name
				})
				drive = filteredReducedDevices[0].Name
			}
		}
	}
	return drive, nil
}

// converts the size of Hetzner drives, e.g. 3,5T to an int, here 3,500,000 (MB)
func convertSizeToInt(str string) (x int, err error) {
	s := str
	var m float64
	var z float64
	strings.ReplaceAll(s, ",", ".")
	if strings.HasSuffix(s, "T") {
		m = 1000000
	} else if strings.HasSuffix(s, "G") {
		m = 1000
	} else if strings.HasSuffix(s, "M") {
		m = 1
	} else {
		return 0, errors.Errorf("Unknown unit in size %s", s)
	}

	s = s[:len(s)-1]

	z, err = strconv.ParseFloat(s, 32)

	if err != nil {
		return 0, errors.Errorf("Error while converting string %s to integer: %s", s, err)
	}
	x = int(z * m)
	return x, nil
}

func (s *Service) retrieveSSHSecret(ctx context.Context) (sshKeyName string, publicKey string, privateKey string, err error) {
	// retrieve token secret
	var tokenSecret corev1.Secret
	tokenSecretName := types.NamespacedName{Namespace: s.scope.HcloudCluster.Namespace, Name: s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.TokenName}
	if err := s.scope.Client.Get(ctx, tokenSecretName, &tokenSecret); err != nil {
		return "", "", "", errors.Errorf("error getting referenced token secret/%s: %s", tokenSecretName, err)
	}

	publicKeyTokenBytes, keyExists := tokenSecret.Data[s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.PublicKey]
	if !keyExists {
		return "", "", "", errors.Errorf("error key %s does not exist in secret/%s", s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.PublicKey, tokenSecretName)
	}
	privateKeyTokenBytes, keyExists := tokenSecret.Data[s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.PrivateKey]
	if !keyExists {
		return "", "", "", errors.Errorf("error key %s does not exist in secret/%s", s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.PrivateKey, tokenSecretName)
	}
	sshKeyNameTokenBytes, keyExists := tokenSecret.Data[s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.SSHKeyName]
	if !keyExists {
		return "", "", "", errors.Errorf("error key %s does not exist in secret/%s", s.scope.HetznerBareMetalMachine.Spec.SSHTokenRef.SSHKeyName, tokenSecretName)
	}

	sshKeyName = string(sshKeyNameTokenBytes)
	privateKey = string(privateKeyTokenBytes)
	publicKey = string(publicKeyTokenBytes)
	return sshKeyName, publicKey, privateKey, nil
}

func runSSH(command, ip string, port int, privateSSHKey string) (stdout string, stderr string, err error) {

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey([]byte(privateSSHKey))
	if err != nil {
		return "", "", errors.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // ssh.FixedHostKey(hostKey),
		Timeout:         15 * time.Second,
	}

	// Connect to the remote server and perform the SSH handshake.
	var client *ssh.Client
	var check bool
	for i := 0; i < (maxWaitingTimeForSSHConnection / 30); i++ {
		client, err = ssh.Dial("tcp", ip+":"+strconv.Itoa(port), config)
		if err != nil {
			// If the SSH connection could not be established, then retry 30 sec later
			time.Sleep(30 * time.Second)
			continue
		}
		check = true
		break
	}

	if check == false {
		return "", "", errors.Errorf("Unable to establish connection to remote server: %s", err)
	}

	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	sess.Stdout = &stdoutBuffer
	sess.Stderr = &stderrBuffer

	err = sess.Run(command)

	stdout = stdoutBuffer.String()
	stderr = stderrBuffer.String()
	return stdout, stderr, err
}
