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

#-----------------------------------------------------------------------------------------------------------------------
# Rules (https://www.gnu.org/software/make/manual/html_node/Rule-Introduction.html#Rule-Introduction)
#-----------------------------------------------------------------------------------------------------------------------
$(GO_BIN)/golangci-lint:
	@echo "==> Installing golangci-lint within "${GO_BIN}""
	@go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest

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
	@go install github.com/golang/mock/mockgen@latest

$(MOCK_SRC_DIR):
	@echo "==> Cloning OpenFGA Go SDK within ${MOCK_SRC_DIR}"
	@git clone https://github.com/openfga/go-sdk ${MOCK_SRC_DIR}

$(MOCK_DIR)/client.go: $(GO_BIN)/mockgen $(MOCK_SRC_DIR)
	@echo "==> Generating client mocks within ${MOCK_DIR}"
	cd ${MOCK_SRC_DIR} && mockgen -source client/client.go -destination ${MOCK_DIR}/client.go

#-----------------------------------------------------------------------------------------------------------------------
# Phony Rules(https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html)
#-----------------------------------------------------------------------------------------------------------------------
.PHONY: build run clean test lint audit format

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the CLI binary
	@echo "==> Building binary within ${BUILD_DIR}/${BINARY_NAME}"
	@go build -v -ldflags "$(GO_LINKER_FLAGS)" -o "${BUILD_DIR}/${BINARY_NAME}" "$(CURDIR)/cmd/fga/main.go"
 
install: ## Install the CLI binary
	@$(MAKE) build BUILD_DIR="$(GO_BIN)"

run: $(GO_BIN)/CompileDaemon ## Watch for changes and recompile the CLI binary
	@echo "==> Watching for changes"
	@CompileDaemon -build='make install' -command='fga --version'
 
clean: ## Clean project files
	@echo "==> Cleaning project files"
	@go clean
	@rm -f ${BUILD_DIR}

test: ## Run tests
	go test -race \
		-coverpkg=./... \
		-coverprofile=coverageunit.tmp.out \
		-covermode=atomic \
		-count=1 \
		-timeout=5m \
		./...
	@cat coverageunit.tmp.out | grep -v "mocks" > coverageunit.out
	@rm coverageunit.tmp.out

lint: $(GO_BIN)/golangci-lint ## Lint Go source files
	@echo "==> Linting Go source files"
	@golangci-lint run -v --fix -c .golangci.yaml ./...

audit: $(GO_BIN)/govulncheck ## Audit Go source files
	@echo "==> Checking Go source files for vulnerabilities"
	govulncheck ./...

format: $(GO_BIN)/gofumpt ## Format Go source files
	@echo "==> Formatting project files"
	gofumpt -w .

mocks: $(MOCK_DIR)/*.go ## Generate Go mocks
	@echo "==> Mocks generated"
