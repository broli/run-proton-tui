BINARY_NAME := rpt
BUILD_DIR := bin
CMD_DIR := ./cmd/rpt
PREFIX ?= $(HOME)/bin

.PHONY: all build clean test lint install

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "[✓] Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test -v ./...

lint:
	go vet ./...

install: build
	@mkdir -p $(PREFIX)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(PREFIX)/$(BINARY_NAME)
	ln -sf $(PREFIX)/$(BINARY_NAME) $(PREFIX)/rptui
	ln -sf $(PREFIX)/$(BINARY_NAME) $(PREFIX)/run-proton
	@echo "[✓] Installed $(PREFIX)/$(BINARY_NAME) and symlinks (rptui, run-proton)"

clean:
	rm -rf $(BUILD_DIR)
