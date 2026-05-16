GO ?= go
BASE := nodora.org/nodora
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
LDFLAGS := -ldflags "-w -s $(if $(VERSION),-X $(BASE).Version=$(VERSION))"

.PHONY: all build clean

all: build

build: lib/nodora.wasm.gz
	@echo "Build complete"

lib/nodora.wasm.gz: lib/nodora.wasm
	gzip -9 -k lib/nodora.wasm

lib/nodora.wasm: export GOOS := js
lib/nodora.wasm: export GOARCH := wasm
lib/nodora.wasm: main.go
	@echo "Building WASM binary..."
	$(GO) build -trimpath $(LDFLAGS) -o $@ .

clean:
	rm -f lib/nodora.wasm*

test: build
	@echo "Running tests..."
	bun test
