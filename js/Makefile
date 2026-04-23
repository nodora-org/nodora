GO ?= go
BASE := nodora.org/nodora
WASM_EXEC_URL = https://raw.githubusercontent.com/golang/go/go1.24.12/lib/wasm/wasm_exec.js
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
LDFLAGS := $(if $(VERSION),-ldflags "-X $(BASE).Version=$(VERSION)",)

.PHONY: all build clean

all: build

build: lib/nodora.wasm lib/wasm_exec.js
	@echo "Build complete"

lib/nodora.wasm: export GOOS := js
lib/nodora.wasm: export GOARCH := wasm
lib/nodora.wasm: main.go
	@echo "Building WASM binary..."
	$(GO) build $(LDFLAGS) -o $@ .

lib/wasm_exec.js:
	@echo "Setting up wasm_exec.js..."
	curl -sL "$(WASM_EXEC_URL)" -o $@

clean:
	rm -f lib/nodora.wasm
	rm -f lib/wasm_exec.js

test: build
	@echo "Running tests..."
	bun test

deploy: build
	@echo "Deploying to dist..."
	rm -rf dist/
	mkdir -p dist
	cp -r lib/ dist/
	cp package.json dist/
	cp README.md dist/
