SKILL_NAME ?= openscad
PREFIX     ?= $(HOME)/.claude/skills
DEST       := $(PREFIX)/$(SKILL_NAME)

SOURCES := SKILL.md references tools

ifeq ($(strip $(SKILL_NAME)),)
$(error SKILL_NAME is empty — refusing to install)
endif
ifeq ($(strip $(PREFIX)),)
$(error PREFIX is empty — refusing to install)
endif

.PHONY: install uninstall

install:
	@for src in $(SOURCES); do \
		if [ ! -e "$$src" ]; then \
			echo "error: $$src not found"; exit 1; \
		fi; \
	done
	rm -rf "$(DEST)"
	mkdir -p "$(DEST)"
	cp -R $(SOURCES) "$(DEST)/"
	@echo "installed -> $(DEST)"

uninstall:
	rm -rf "$(DEST)"
	@echo "removed -> $(DEST)"
