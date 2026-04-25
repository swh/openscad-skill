SKILL_NAME ?= openscad-bosl2
PREFIX     ?= $(HOME)/.claude/skills
DEST       := $(PREFIX)/$(SKILL_NAME)

GO         ?= go
SKILL_DIR  := openscad-bosl2
TOOLS_DIR  := $(SKILL_DIR)/tools
BINS       := $(TOOLS_DIR)/openscad-pack-3mf $(TOOLS_DIR)/openscad-build-template

ifeq ($(strip $(SKILL_NAME)),)
$(error SKILL_NAME is empty — refusing to install)
endif
ifeq ($(strip $(PREFIX)),)
$(error PREFIX is empty — refusing to install)
endif

.PHONY: all build install uninstall clean test

all: build

build: $(BINS)

$(TOOLS_DIR)/openscad-pack-3mf: $(shell find cmd/openscad-pack-3mf internal -name '*.go' 2>/dev/null)
	$(GO) build -o $@ ./cmd/openscad-pack-3mf

$(TOOLS_DIR)/openscad-build-template: $(shell find cmd/openscad-build-template internal -name '*.go' 2>/dev/null)
	$(GO) build -o $@ ./cmd/openscad-build-template

# Install copies the contents of the skill subdir (binaries included) into
# the destination, so $(DEST)/SKILL.md and $(DEST)/tools/openscad-pack-3mf.
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

test:
	$(GO) test ./...
	@echo "smoke: pack-3mf default"
	$(TOOLS_DIR)/openscad-pack-3mf -o /tmp/seed-packed.3mf $(TOOLS_DIR)/template-seed.scad
