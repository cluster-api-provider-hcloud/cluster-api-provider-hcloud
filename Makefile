# This makefile provider some wrapper around bazel targets

# from https://suva.sh/posts/well-documented-makefiles/
.PHONY: help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run all tests through bazel
	bazel test //...

.PHONY: run
run: ## Run controller's binary
	bazel run //cmd/cluster-api-provider-hcloud:run

.PHONY: install
install: manifests ## Install CRDs into a cluster
	kubectl apply -k config/crd

.PHONY: install_all
install_all: manifests ## Install CRDs into a cluster

.PHONY: manifests
manifests: ## Update generated manifests
	bazel run //hack:update-crds

.PHONY: deploy_kind
deploy_kind: ## Deploy latest image and manifests to a kind cluster
	bazel run //cmd/cluster-api-provider-hcloud:deploy
	kubectl wait -n capi-webhook-system deployment capi-controller-manager --for=condition=Available --timeout=120s
	kubectl apply -f demo/ClusterResourceSets
# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# TODO: Bazelify
mockgen:
	mkdir -p pkg/cloud/scope/mock
	mockgen github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope HcloudClient,Manifests,Packer > pkg/cloud/scope/mock/scope.go


# Generate hack/build/repos.bzl from go.mod
.PHONY: bazel_repos
bazel_repos:
	bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=repositories.bzl%go_repositories -prune=true

.PHONY: delete_capihc
delete_capihc:
	kubectl delete namespace capi-hcloud-system
