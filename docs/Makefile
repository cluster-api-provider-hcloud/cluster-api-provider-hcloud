MDBOOK_VERSION := 0.3.7
MDBOOK_SHA256 := 9b61fe2476dfd8b8459702966bf2f226f9d6eddc00133adf09e37d30e7eab117
MDBOOK := bin/mdbook-$(MDBOOK_VERSION)

MDBOOK_MERMAID_VERSION := 0.2.2
MDBOOK_MERMAID_SHA256 := bd31bc95a89ef3e0178bc081645313e2122e00a41b35cf3fb5b3474a9d4e40f5 
MDBOOK_MERMAID := bin/mdbook-mermaid-$(MDBOOK_MERMAID_VERSION)

PATH := $(CURDIR)/bin:$(PATH)

# from https://suva.sh/posts/well-documented-makefiles/
.PHONY: help
 help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

$(MDBOOK):
	mkdir -p bin
	curl -Lo $@.tar.gz "https://github.com/rust-lang/mdBook/releases/download/v$(MDBOOK_VERSION)/mdbook-v$(MDBOOK_VERSION)-x86_64-unknown-linux-gnu.tar.gz"
	echo "$(MDBOOK_SHA256) $@.tar.gz" | sha256sum -c
	tar xfz $@.tar.gz mdbook
	mv mdbook $@
	rm $@.tar.gz
	ln -fs $(shell basename $@) bin/mdbook

$(MDBOOK_MERMAID):
	mkdir -p bin
	curl -Lo $@.tar.gz "https://github.com/badboy/mdbook-mermaid/releases/download/$(MDBOOK_MERMAID_VERSION)/mdbook-mermaid-$(MDBOOK_MERMAID_VERSION)-x86_64-unknown-linux-gnu.tar.gz"
	echo "$(MDBOOK_MERMAID_SHA256) $@.tar.gz" | sha256sum -c
	tar xfz $@.tar.gz mdbook-mermaid
	mv mdbook-mermaid $@
	rm $@.tar.gz
	ln -fs $(shell basename $@) bin/mdbook-mermaid

.PHONY: generate
generate: $(MDBOOK) $(MDBOOK_MERMAID) ## Build the static website
	$(MDBOOK) build

.PHONY: serve
serve: ## Serve the docs on localhost
	$(MDBOOK) serve

.PHONY: clean
clean: ## Remove build artificats
	rm -rf book
