fetch: hcloud-ip-floater/manifests.json hcloud-cloud-controller-manager/manifests.json calico/manifests.json hcloud-csi/manifests.json

HCLOUD_IP_FLOATER_VERSION := 0.1.4
hcloud-ip-floater/manifests.json:
	kubectl convert -k hcloud-ip-floater/upstream -o json | jq .items > $@

hcloud-metallb-floater/manifests.json:
	kubectl convert -k hcloud-metallb-floater/upstream -o json | jq .items > $@

HCLOUD_CLOUD_CONTROLLER_MANAGER_VERSION := 1.5.0
hcloud-cloud-controller-manager/manifests.json:
	mkdir -p $(shell dirname "$@")
	curl -sL https://raw.githubusercontent.com/hetznercloud/hcloud-cloud-controller-manager/v$(HCLOUD_CLOUD_CONTROLLER_MANAGER_VERSION)/deploy/v$(HCLOUD_CLOUD_CONTROLLER_MANAGER_VERSION).yaml | \
		go run ../hack/yaml2json.go | \
		jq -s  > $@

HCLOUD_CSI_VERSION := 1.2.2
hcloud-csi/manifests.json:
	mkdir -p $(shell dirname "$@")
	curl -sL https://raw.githubusercontent.com/hetznercloud/csi-driver/v$(HCLOUD_CSI_VERSION)/deploy/kubernetes/hcloud-csi.yml | \
		go run ../hack/yaml2json.go | \
		jq -s  > $@

CALICO_VERSION := 3.13
calico/manifests.json:
	mkdir -p $(shell dirname "$@")
	curl -sL https://docs.projectcalico.org/v$(CALICO_VERSION)/manifests/calico.yaml | \
		go run ../hack/yaml2json.go | \
		jq -s  > $@

METALLB_VERSION := 0.9.3
metallb/manifests.json:
	mkdir -p $(shell dirname "$@")
	curl -sL https://github.com/metallb/metallb/raw/v0.9.3/manifests/metallb.yaml | \
		go run ../hack/yaml2json.go | \
		jq -s  > $@

METRICS_SERVER_VERSION := 0.3.6
metrics-server/manifests.json:
	mkdir -p $(shell dirname "$@")/tmp
	curl -sL https://github.com/kubernetes-sigs/metrics-server/archive/v$(METRICS_SERVER_VERSION).tar.gz | tar xzf - --strip-components 3 -C $(shell dirname "$@")/tmp metrics-server-$(METRICS_SERVER_VERSION)/deploy/1.8+
	cat $(shell dirname "$@")/tmp/*.yaml | go run ../hack/yaml2json.go | jq -s  > $@
	rm -r $(shell dirname "$@")/tmp
