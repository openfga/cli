#-----------------------------------------------------------------------------------------------------------------------
# Variables (https://www.gnu.org/software/make/manual/html_node/Using-Variables.html#Using-Variables)
#-----------------------------------------------------------------------------------------------------------------------
BINARY_NAME = fga
BUILD_DIR ?= $(CURDIR)/dist
GO_BIN ?= $(shell go env GOPATH)/bin

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

$(BUILD_DIR)/$(BINARY_NAME):
	@echo "==> Building binary within ${BUILD_DIR}/${BINARY_NAME}"
	@go build -v -o ${BUILD_DIR}/${BINARY_NAME} main.go

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

build:
	@echo "==> Building binary within ${BUILD_DIR}/${BINARY_NAME}"
	@go build -v -o ${BUILD_DIR}/${BINARY_NAME} main.go
 
run: build
	${BUILD_DIR}/${BINARY_NAME} $(ARGS)
 
clean:
	@echo "==> Cleaning project files"
	go clean
	rm -f ${BUILD_DIR}

test:
	go test -race \
			-coverpkg=./... \
			-coverprofile=coverageunit.tmp.out \
			-covermode=atomic \
			-count=1 \
			-timeout=5m \
			./...
	@cat coverageunit.tmp.out | grep -v "mocks" > coverageunit.out
	@rm coverageunit.tmp.out

lint: $(GO_BIN)/golangci-lint
	@echo "==> Linting Go source files"
	@golangci-lint run -v --fix -c .golangci.yaml ./...

audit: $(GO_BIN)/govulncheck
	@echo "==> Checking Go source files for vulnerabilities"
	govulncheck ./...

format: $(GO_BIN)/gofumpt $(GO_BIN)/gci
	@echo "==> Formatting project files"
	gofumpt -w .

mocks: $(MOCK_DIR)/*.go
	@echo "==> Mocks generated"
