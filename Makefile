BINARY_NAME := rpt
BUILD_DIR := bin
CMD_DIR := ./cmd/rpt
PREFIX ?= $(HOME)/bin
MAN_DIR ?= $(HOME)/.local/share/man/man1

.PHONY: all build clean test lint install install-man

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "[✓] Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test -v ./...

lint:
	go vet ./...

install-man:
	@mkdir -p $(MAN_DIR)
	install -m 644 man/rpt.1 $(MAN_DIR)/rpt.1
	@echo "[✓] Installed man page to $(MAN_DIR)/rpt.1"

install: build install-man
	@mkdir -p $(PREFIX)
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(PREFIX)/$(BINARY_NAME)
	ln -sf $(PREFIX)/$(BINARY_NAME) $(PREFIX)/rptui
	ln -sf $(PREFIX)/$(BINARY_NAME) $(PREFIX)/run-proton
	@mkdir -p $(HOME)/.config/fish/completions
	install -m 644 completions/rpt.fish $(HOME)/.config/fish/completions/rpt.fish
	ln -sf $(HOME)/.config/fish/completions/rpt.fish $(HOME)/.config/fish/completions/run-proton.fish
	ln -sf $(HOME)/.config/fish/completions/rpt.fish $(HOME)/.config/fish/completions/rptui.fish
	@echo "[✓] Installed $(PREFIX)/$(BINARY_NAME), symlinks, man page, and completions"

clean:
	rm -rf $(BUILD_DIR)
