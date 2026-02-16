# Nodora WASM Build Configuration

GO ?= go
WASM_EXEC_URL = https://raw.githubusercontent.com/golang/go/go1.24.12/lib/wasm/wasm_exec.js

.PHONY: all build clean

all: build

build: lib/nodora.wasm lib/wasm_exec.js
	@echo "Build complete"

lib/nodora.wasm: main.go
	@echo "Building WASM binary..."
	GOOS=js GOARCH=wasm $(GO) build -o $@ .

lib/wasm_exec.js:
	@echo "Setting up wasm_exec.js..."
	curl -sL "$(WASM_EXEC_URL)" -o $@

clean:
	rm -f lib/nodora.wasm
	rm -f lib/wasm_exec.js

test: build
	@echo "Running tests..."
	bun test

# Development helper to copy wasm to a target directory
deploy: build
	@echo "Deploying to dist..."
	rm -rf dist/
	mkdir -p dist
	cp -r lib/ dist/
	cp package.json dist/
	cp README.md dist/
