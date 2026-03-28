VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/joelazar/kagi/internal/version.Version=$(VERSION)"

.PHONY: build test lint fmt clean install

build:
	go build $(LDFLAGS) -o bin/kagi ./cmd/kagi

test:
	go test ./...

test-integration:
	go test -tags integration ./...

lint:
	golangci-lint run

fmt:
	gofumpt -w -l .

clean:
	rm -rf bin/

install:
	go install $(LDFLAGS) ./cmd/kagi
