.PHONY: build install test clean

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X github.com/artisanscompany/crewsbase-cli/internal/cmd.Version=$(VERSION) -X github.com/artisanscompany/crewsbase-cli/internal/cmd.Commit=$(COMMIT) -X github.com/artisanscompany/crewsbase-cli/internal/api.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/crewsbase ./cmd/crewsbase

install:
	go install $(LDFLAGS) ./cmd/crewsbase

test:
	go test ./...

clean:
	rm -rf bin/
