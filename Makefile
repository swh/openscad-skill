SKILL_NAME ?= openscad-bosl2
PREFIX     ?= $(HOME)/.claude/skills
DEST       := $(PREFIX)/$(SKILL_NAME)

GO         ?= go
SKILL_DIR  := openscad-bosl2
TOOLS_DIR  := $(SKILL_DIR)/tools
BINS       := $(TOOLS_DIR)/openscad-pack-3mf $(TOOLS_DIR)/openscad-build-template

# Real entry-point logic lives in internal/cli/<name>; cmd/ shims that
# `go build` requires are generated at build time and removed afterwards,
# so cmd/ doesn't need to be tracked in git.
GENERATED_CMD := cmd

ifeq ($(strip $(SKILL_NAME)),)
$(error SKILL_NAME is empty — refusing to install)
endif
ifeq ($(strip $(PREFIX)),)
$(error PREFIX is empty — refusing to install)
endif

.PHONY: all build install uninstall clean test

all: build

build: $(BINS)

# `define / endef` keeps the multiline shim readable. The shim is a 1-line
# main package whose only job is to call <pkg>.Main().
define SHIM_TEMPLATE
package main

import "github.com/swh/openscad-skill/internal/cli/$(2)"

func main() { $(2).Main() }
endef

$(TOOLS_DIR)/openscad-pack-3mf: $(shell find internal -name '*.go' 2>/dev/null)
	@mkdir -p $(GENERATED_CMD)/openscad-pack-3mf
	@printf 'package main\n\nimport "github.com/swh/openscad-skill/internal/pack3mf"\n\nfunc main() { pack3mf.Main() }\n' > $(GENERATED_CMD)/openscad-pack-3mf/main.go
	$(GO) build -o $@ ./$(GENERATED_CMD)/openscad-pack-3mf
	@rm -rf $(GENERATED_CMD)/openscad-pack-3mf
	@rmdir $(GENERATED_CMD) 2>/dev/null || true

$(TOOLS_DIR)/openscad-build-template: $(shell find internal -name '*.go' 2>/dev/null)
	@mkdir -p $(GENERATED_CMD)/openscad-build-template
	@printf 'package main\n\nimport "github.com/swh/openscad-skill/internal/buildtemplate"\n\nfunc main() { buildtemplate.Main() }\n' > $(GENERATED_CMD)/openscad-build-template/main.go
	$(GO) build -o $@ ./$(GENERATED_CMD)/openscad-build-template
	@rm -rf $(GENERATED_CMD)/openscad-build-template
	@rmdir $(GENERATED_CMD) 2>/dev/null || true

install: build
	rm -rf "$(DEST)"
	mkdir -p "$(DEST)"
	cp -R $(SKILL_DIR)/. "$(DEST)/"
	@echo "installed -> $(DEST)"

uninstall:
	rm -rf "$(DEST)"
	@echo "removed -> $(DEST)"

clean:
	rm -f $(BINS)
	rm -rf $(GENERATED_CMD)

test:
	$(GO) test ./...
	@echo "smoke: pack-3mf default"
	$(TOOLS_DIR)/openscad-pack-3mf -o /tmp/seed-packed.3mf $(TOOLS_DIR)/template-seed.scad
