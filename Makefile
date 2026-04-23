GO := go

BASE := nodora.org/nodora
MAIN_PACKAGE := ./cmd/nodora
BUILD_DIR := build

VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
LDFLAGS := $(if $(VERSION),-ldflags "-X $(BASE).Version=$(VERSION)",)

build:
	@echo Building...
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR) $(MAIN_PACKAGE)

clean:
	@echo Cleaning...
	@rm -rf $(BUILD_DIR)

test:
	go test -count=1 ./...

.PHONY: build clean