fetch: hcloud-ip-floater/manifests.json hcloud-cloud-controller-manager/manifests.json calico/manifests.json

HCLOUD_IP_FLOATER_VERSION := 0.1.4
hcloud-ip-floater/manifests.json:
	kubectl convert -k hcloud-ip-floater/upstream -o json | jq .items > $@

HCLOUD_CLOUD_CONTROLLER_MANAGER_VERSION := 1.5.0
hcloud-cloud-controller-manager/manifests.json:
	mkdir -p $(shell dirname "$@")
	curl https://raw.githubusercontent.com/hetznercloud/hcloud-cloud-controller-manager/v$(HCLOUD_CLOUD_CONTROLLER_MANAGER_VERSION)/deploy/v$(HCLOUD_CLOUD_CONTROLLER_MANAGER_VERSION).yaml | go run ../hack/yaml2json.go | jq -s  > $@

CALICO_VERSION := 3.11
calico/manifests.json:
	mkdir -p $(shell dirname "$@")
	curl https://docs.projectcalico.org/v$(CALICO_VERSION)/manifests/calico.yaml  | go run ../hack/yaml2json.go | jq -s  > $@
