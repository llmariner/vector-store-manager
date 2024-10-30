.PHONY: default
default: test

include common.mk

.PHONY: test
test: go-test-all

.PHONY: test-integration
test-integration:
	@go test ./... -tags=integration

.PHONY: lint
lint: go-lint-all helm-lint git-clean-check

.PHONY: generate
generate: buf-generate-all typescript-compile

.PHONY: build-server
build-server:
	go build -o ./bin/server ./server/cmd/

.PHONY: build-docker-server
build-docker-server:
	docker build --build-arg TARGETARCH=amd64 -t llmariner/vector-store-manager-server:latest -f build/server/Dockerfile .

.PHONY: check-helm-tool
check-helm-tool:
	@command -v helm-tool >/dev/null 2>&1 || $(MAKE) install-helm-tool

.PHONY: install-helm-tool
install-helm-tool:
	go install github.com/cert-manager/helm-tool@latest

.PHONY: generate-chart-schema
generate-chart-schema: check-helm-tool
	@cd ./deployments/server && helm-tool schema > values.schema.json

.PHONY: helm-lint
helm-lint: generate-chart-schema
	cd ./deployments/server && helm-tool lint
	helm lint ./deployments/server
