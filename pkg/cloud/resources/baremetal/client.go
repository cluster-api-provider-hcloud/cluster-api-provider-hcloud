package baremetal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type BareMetalServerStatus struct {
	IPv4       string `json:"server_ip,omitempty"`
	ServerID   int    `json:"server_number,omitempty"`
	ServerName string `json:"server_name,omitempty"`
	Status     string `json:"status,omitempty"`
	Cancelled  bool   `json:"cancelled,omitempty"`
	IPv6       string `json:"ipv6,omitempty"`
}

type GeneralServerData struct {
	ServerIP   string `json:"server_ip,omitempty"`
	ServerName string `json:"server_name,omitempty"`
}

func ResetServer(serverIP, resetType, userName, password string) error {
	data := url.Values{}
	data.Set("type", resetType)

	req, err := NewHTTPRequestWithAuth(
		"POST",
		"https://robot-ws.your-server.de/reset/"+serverIP,
		strings.NewReader(data.Encode()),
		userName, password)
	if err != nil {
		return errors.Errorf("unable to create http request: %v", err)
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Errorf("error while executing http request: %v", err)
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("HTTP request gave the error code %s", resp.Status)
	}

	return nil
}

func ActivateRescue(serverIP, sshFingerprint, userName, password string) error {

	data := url.Values{}
	data.Set("os", "linux")
	data.Set("authorized_key", sshFingerprint)

	req, err := NewHTTPRequestWithAuth(
		"POST",
		"https://robot-ws.your-server.de/boot/"+serverIP+"/rescue",
		strings.NewReader(data.Encode()),
		userName, password)
	if err != nil {
		return errors.Errorf("unable to create http request: %v", err)
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Errorf("error while executing http request: %v", err)
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("HTTP request gave the error code %s", resp.Status)
	}

	return nil
}

func GetSSHFingerprintFromName(name, userName, password string) (fingerprint string, err error) {

	req, err := NewHTTPRequestWithAuth("GET", "https://robot-ws.your-server.de/key/", nil, userName, password)
	if err != nil {
		return "", errors.Errorf("unable to create http request: %v", err)
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", errors.Errorf("error while executing http request: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", errors.Errorf("HTTP request gave the error code %s", resp.Status)
	}

	type hetznerSSHKeySpec struct {
		Name        string `json:"name"`
		Fingerprint string `json:"fingerprint"`
	}

	type hetznerSSHKey struct {
		Key hetznerSSHKeySpec `json:"key"`
	}

	var sshKeys []hetznerSSHKey

	err = json.NewDecoder(resp.Body).Decode(&sshKeys)
	if err != nil {
		return "", errors.Errorf("unable to decode response body: %v", err)
	}

	if len(sshKeys) == 0 {
		return "", errors.New("No SSH Keys given")
	}

	var found bool
	for _, key := range sshKeys {
		if name == key.Key.Name {
			fingerprint = key.Key.Fingerprint
			found = true
		}
	}

	if found == false {
		return "", errors.Errorf("No SSH key with name %s found", name)
	}

	return fingerprint, nil
}

func ChangeBareMetalServerName(serverIP, name, userName, password string) error {

	data := url.Values{}
	data.Set("server_name", name)

	req, err := NewHTTPRequestWithAuth("POST", "https://robot-ws.your-server.de/server/"+serverIP, strings.NewReader(data.Encode()), userName, password)
	if err != nil {
		return errors.Errorf("unable to create http request: %v", err)
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Errorf("error while executing http request: %v", err)
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("HTTP request gave the error code %s", resp.Status)
	}

	return nil
}

func GetBareMetalServer(serverIP, userName, password string) (*infrav1.BareMetalMachineStatus, error) {

	type subnet struct {
		IP   string `json:"ip,omitempty"`
		Mask string `json:"mask,omitempty"`
	}

	type hetznerServerStatus struct {
		IPv4       string   `json:"server_ip"`
		ServerID   int      `json:"server_number"`
		ServerName string   `json:"server_name"`
		Status     string   `json:"status"`
		Cancelled  bool     `json:"cancelled"`
		Subnet     []subnet `json:"subnet"`
		Reset      bool     `json:"reset"`
		Rescue     bool     `json:"rescue"`
	}

	type hetznerServer struct {
		Server *hetznerServerStatus `json:"server,omitempty"`
	}

	req, err := NewHTTPRequestWithAuth("GET", "https://robot-ws.your-server.de/server/"+serverIP, nil, userName, password)
	if err != nil {
		return nil, errors.Errorf("unable to create http request: %v", err)
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Errorf("error while executing http request: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("HTTP request gave the error code %s", resp.Status)
	}

	var server hetznerServer

	err = json.NewDecoder(resp.Body).Decode(&server)
	if err != nil {
		return nil, errors.Errorf("unable to decode response body: %v", err)
	}

	if server.Server == nil {
		return nil, errors.Errorf("Could not find bare metal server with IP %s", serverIP)
	}
	var status infrav1.BareMetalMachineStatus

	status.IPv4 = server.Server.IPv4
	status.ServerID = server.Server.ServerID
	status.ServerName = server.Server.ServerName
	status.Status = server.Server.Status
	status.Cancelled = server.Server.Cancelled
	status.Reset = server.Server.Reset
	status.Rescue = server.Server.Rescue

	if len(server.Server.Subnet) > 0 {
		status.IPv6 = server.Server.Subnet[0].IP + "2"
	} else {
		return nil, errors.New("No Subnet specified for bare metal server")
	}

	return &status, nil
}

func ListBareMetalServers(userName, password string) (serverList []GeneralServerData, err error) {
	req, err := NewHTTPRequestWithAuth("GET", "https://robot-ws.your-server.de/server", nil, userName, password)
	if err != nil {
		return nil, errors.Errorf("unable to create http request: %v", err)
	}
	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Errorf("error while executing http request: %v", err)
	}

	type hetznerServer struct {
		Server GeneralServerData `json:"server,omitempty"`
	}

	var servers []hetznerServer

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("HTTP request gave the error code %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&servers)
	if err != nil {
		return nil, errors.Errorf("unable to decode response body: %v", err)
	}

	for _, server := range servers {
		serverList = append(serverList, server.Server)
	}

	return serverList, nil
}

func NewHTTPRequestWithAuth(method, url string, data io.Reader, userName, password string) (req *http.Request, err error) {

	if method == "POST" {

		if data == nil {
			return nil, errors.New("No data specified for POST request")
		}

		req, err = http.NewRequest(method, url, data)
		if err != nil {
			return nil, errors.Errorf("unable to create http request: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	} else if method == "GET" {

		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, errors.Errorf("unable to create http request: %v", err)
		}
	} else {
		return nil, errors.Errorf("Unknown http method %s", method)
	}

	req.SetBasicAuth(userName, password)

	return req, nil
}

func runSSH(command, ip string, port int, privateSSHKey string, maxTime int) (stdout string, stderr string, err error) {

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
	}

	// Connect to the remote server and perform the SSH handshake.
	var client *ssh.Client
	var check bool
	for i := 0; i < (maxTime / 15); i++ {
		client, err = ssh.Dial("tcp", ip+":"+strconv.Itoa(port), config)
		if err != nil {
			// If the SSH connection could not be established, then retry 15 sec later
			time.Sleep(15 * time.Second)
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
