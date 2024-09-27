.PHONY: default
default: test

include common.mk

.PHONY: test
test: go-test-all

.PHONY: test-integration
test-integration:
	@go test ./... -tags=integration

.PHONY: lint
lint: go-lint-all git-clean-check

.PHONY: generate
generate: buf-generate-all typescript-compile

.PHONY: build-server
build-server:
	go build -o ./bin/server ./server/cmd/

.PHONY: build-docker-server
build-docker-server:
	docker build --build-arg TARGETARCH=amd64 -t llmariner/vector-store-manager-server:latest -f build/server/Dockerfile .
