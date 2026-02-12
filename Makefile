GO := go

BASE := nodora.org/nodora
MAIN_PACKAGE := ./cmd/nodora
BUILD_DIR := build

ifeq ($(OS),Windows_NT)
	NULLDEV = NUL
	MKDIR_BUILD = if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	RM_BUILD = if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
else
	NULLDEV = /dev/null
	MKDIR_BUILD = mkdir -p $(BUILD_DIR)
	RM_BUILD = rm -rf $(BUILD_DIR)
endif

VERSION := $(shell git describe --tags --exact-match HEAD 2> $(NULLDEV))
ifeq ($(VERSION),)
	LDFLAGS :=
else
	LDFLAGS := -ldflags "-X $(BASE)/version.Version=$(VERSION)"
endif

build:
	@echo Building...
	@$(MKDIR_BUILD)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR) $(MAIN_PACKAGE)

clean:
	@echo Cleaning...
	@$(RM_BUILD)

.PHONY: build clean