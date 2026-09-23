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

publish-wiki:
	@echo "Publishing documentation from docs/wiki to GitHub Wiki..."
	@rm -rf /tmp/rpt-wiki
	@git clone https://github.com/broli/run-proton-tui.wiki.git /tmp/rpt-wiki || { \
		echo ""; \
		echo "[-] Could not clone GitHub Wiki repository."; \
		echo "    Note: GitHub only creates the wiki repository AFTER you create the first page via the web UI."; \
		echo "    1. Visit https://github.com/broli/run-proton-tui/wiki"; \
		echo "    2. Click 'Create the first page' and click 'Save Page'"; \
		echo "    3. Run 'make publish-wiki' again."; \
		exit 1; \
	}
	@cp -f docs/wiki/*.md /tmp/rpt-wiki/
	@cd /tmp/rpt-wiki && git add . && git commit -m "docs(wiki): update wiki documentation" && git push origin
	@rm -rf /tmp/rpt-wiki
	@echo "[✓] Successfully published docs/wiki to GitHub Wiki!"
