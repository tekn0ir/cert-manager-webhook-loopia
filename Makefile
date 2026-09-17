IMAGE_NAME := "ghcr.io/tekn0ir/cert-manager-webhook-loopia"
IMAGE_TAG := "latest"

CHART_NAME := "cert-manager-webhook-loopia"
CHART_DIR := "charts/$(CHART_NAME)"

OUT ?= .

# Run the conformance suite, it needs the envtest binaries and a zone you control at Loopia, see the readme.
test:
	eval "$$(sh ./testdata/scripts/fetch-test-binaries.sh --env)" && go test -tags conformance -v .

# Run everything that needs no credentials, the same checks CI runs.
check:
	go vet ./...
	go vet -tags conformance ./...
	go test -v ./...

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

clean: clean-test

clean-test:
	# setup-envtest extracts the binaries read-only, the download directory has to be made writable to remove it.
	chmod -R u+w testdata/bin 2>/dev/null || true
	rm -rf testdata/bin

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template $(CHART_NAME) $(CHART_DIR) \
	    --set image.repository=$(IMAGE_NAME) \
	    --set image.tag=$(IMAGE_TAG) > "$(OUT)/rendered-manifest.yaml"
