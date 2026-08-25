# hcflow Makefile
GO = $(shell which go || echo /media/hickstein/data1/go/bin/go)
BINARY = bin/hcflow
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -ldflags="-X github.com/hickstein/hcflow/cmd/hcflow/commands.Version=$(VERSION)"

.PHONY: build test lint clean install doctor

build:
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o $(BINARY) ./cmd/hcflow

test:
	$(GO) test ./... -v

lint:
	$(GO) vet ./...

install: build
	cp $(BINARY) $(HOME)/.local/bin/hcflow
	@echo "Installed hcflow to ~/.local/bin/hcflow"

clean:
	rm -rf bin/

doctor: build
	./$(BINARY) doctor

.DEFAULT_GOAL := build
