BINARY_NAME=fga

all: build

build:
	go build -o ${BINARY_NAME} main.go
 
run: build
	./${BINARY_NAME}
 
clean:
	go clean
	rm -f ${BINARY_NAME}

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

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && golangci-lint run

audit:
	go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...;

format:
	go install mvdan.cc/gofumpt@latest && gofumpt -w .
	go install github.com/daixiang0/gci@latest && gci write -s standard -s default .
