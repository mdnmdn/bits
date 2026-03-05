BINARY_NAME=cg
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X coingecko-cli/cmd.version=$(VERSION)"

.PHONY: build test lint clean

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	go test -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME)
