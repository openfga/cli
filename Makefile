#-----------------------------------------------------------------------------------------------------------------------
# Variables (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
#-----------------------------------------------------------------------------------------------------------------------
.DEFAULT_GOAL := help

BINARY_NAME = fga
BUILD_DIR ?= $(CURDIR)/dist
GO_BIN ?= $(shell go env GOPATH)/bin
GO_PKG := github.com/openfga/cli

BUILD_INFO_PKG := $(GO_PKG)/internal/build
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GO_LINKER_FLAGS = -X '$(BUILD_INFO_PKG).Version=dev' \
					 -X '$(BUILD_INFO_PKG).Commit=$(GIT_COMMIT)' \
					 -X '$(BUILD_INFO_PKG).Date=$(BUILD_TIME)'

MOCK_DIR ?= internal/mocks
MOCK_SRC_DIR ?= mocks

#-----------------------------------------------------------------------------------------------------------------------
# Rules (https://www.gnu.org/software/make/manual/html_node/Rule-Introduction.html#Rule-Introduction)
#-----------------------------------------------------------------------------------------------------------------------
$(GO_BIN)/golangci-lint:
	@echo "==> Installing golangci-lint within "${GO_BIN}""
	@go install -v github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

$(GO_BIN)/govulncheck:
	@echo "==> Installing govulncheck within "${GO_BIN}""
	@go install -v golang.org/x/vuln/cmd/govulncheck@latest

$(GO_BIN)/gofumpt:
	@echo "==> Installing gofumpt within "${GO_BIN}""
	@go install -v mvdan.cc/gofumpt@latest

$(GO_BIN)/CompileDaemon:
	@echo "==> Installing CompileDaemon within "${GO_BIN}""
	@go install -v github.com/githubnemo/CompileDaemon@latest

$(GO_BIN)/mockgen:
	@echo "==> Installing mockgen within ${GO_BIN}"
	@go install go.uber.org/mock/mockgen@latest

$(GO_BIN)/commander:
	@echo "==> Installing commander within ${GO_BIN}"
	@go install github.com/commander-cli/commander/v2/cmd/commander@latest

$(MOCK_SRC_DIR):
	@echo "==> Cloning OpenFGA Go SDK within ${MOCK_SRC_DIR}"
	@git clone https://github.com/openfga/go-sdk ${MOCK_SRC_DIR}

$(MOCK_DIR)/client.go: $(GO_BIN)/mockgen $(MOCK_SRC_DIR)
	@echo "==> Generating client mocks within ${MOCK_DIR}"
	mockgen -source $(MOCK_SRC_DIR)/client/client.go -destination ${MOCK_DIR}/client.go

#-----------------------------------------------------------------------------------------------------------------------
# Phony Rules(https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: build run clean test lint audit format

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the CLI binary
	@echo "==> Building binary within ${BUILD_DIR}/${BINARY_NAME}"
	@go build -v -ldflags "$(GO_LINKER_FLAGS)" -o "${BUILD_DIR}/${BINARY_NAME}" "$(CURDIR)/cmd/fga/main.go"

build-with-cover: ## Build the CLI binary for the native platform with coverage support
	@echo "Building the cli binary"
	@go build -cover -ldflags "$(GO_LINKER_FLAGS)" -o "${BUILD_DIR}/${BINARY_NAME}" "$(CURDIR)/cmd/fga/main.go"
 
install: ## Install the CLI binary
	@$(MAKE) build BUILD_DIR="$(GO_BIN)"

install-with-cover: ## Install the CLI binary for the native platform with coverage support
	@echo "Installing the CLI binary with coverage support"
	@$(MAKE) build-with-cover BUILD_DIR="$(GO_BIN)"

run: $(GO_BIN)/CompileDaemon ## Watch for changes and recompile the CLI binary
	@echo "==> Watching for changes"
	@CompileDaemon -build='make install' -command='fga --version'
 
clean: ## Clean project files
	@echo "==> Cleaning project files"
	@go clean
	@rm -f ${BUILD_DIR}

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test -race \
		-coverpkg=./... \
		-coverprofile=coverageunit.tmp.out \
		-covermode=atomic \
		-count=1 \
		-timeout=5m \
		./...
	@cat coverageunit.tmp.out | grep -v "mocks" > coverageunit.out
	@rm coverageunit.tmp.out

test-integration: install-with-cover $(GO_BIN)/fga $(GO_BIN)/commander ## Run integration tests
	@echo "==> Running integration tests"
	@mkdir -p "coverage"
	@PATH=$(GO_BIN):$$PATH GOCOVERDIR=coverage bash ./tests/scripts/run-test-suites.sh
	@go tool covdata textfmt -i "coverage" -o "coverage-integration-tests.out"

lint: $(GO_BIN)/golangci-lint ## Lint Go source files
	@echo "==> Linting Go source files"
	@golangci-lint run -v --fix -c .golangci.yaml ./...

audit: $(GO_BIN)/govulncheck ## Audit Go source files
	@echo "==> Checking Go source files for vulnerabilities"
	govulncheck ./...

format: $(GO_BIN)/gofumpt ## Format Go source files
	@echo "==> Formatting project files"
	gofumpt -w .

generate-mocks: $(MOCK_DIR)/*.go ## Generate Go mocks
	@echo "==> Mocks generated"
