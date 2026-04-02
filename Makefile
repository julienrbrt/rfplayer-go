#!/usr/bin/make -f

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

PWD=$(shell pwd)

build:
	go build $(LDFLAGS) -o rfplayer-bin ./cmd/rfplayer/...

install:
	go install $(LDFLAGS) ./cmd/rfplayer/...

test:
	go test -cover ./...

lint:
	go vet ./...

.PHONY: build install test lint
