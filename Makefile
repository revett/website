TAILWIND_VERSION := v4.3.0

UNAME := $(shell uname -s)-$(shell uname -m)
ifeq ($(UNAME),Darwin-arm64)
	TAILWIND_TARGET := macos-arm64
else ifeq ($(UNAME),Darwin-x86_64)
	TAILWIND_TARGET := macos-x64
else
	TAILWIND_TARGET := linux-x64
endif

.PHONY: build serve check thumbs clean

## build: render the full site into dist/
build: bin/tailwindcss
	go run ./generator build
	bin/tailwindcss -i web/css/input.css -o dist/style.css --minify

## serve: local dev server on :8080 with live reload
serve: bin/tailwindcss
	go run ./generator serve

## check: validate internal links in dist/
check:
	go run ./generator check

## thumbs: regenerate the committed 320px thumbnails for post covers (macOS sips)
thumbs:
	mkdir -p content/images/thumbs
	for f in $$(grep -h '^cover:' content/posts/*.md | awk '{print $$2}'); do \
		sips -Z 320 "content$$f" --out "content/images/thumbs/$$(basename $$f)" >/dev/null; \
	done

bin/tailwindcss:
	mkdir -p bin
	curl -sfL -o $@ https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$(TAILWIND_TARGET)
	chmod +x $@

clean:
	rm -rf dist
