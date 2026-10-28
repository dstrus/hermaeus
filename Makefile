# Building Hermaeus needs Go and Node. Running it needs neither: the
# program it builds has the frontend embedded.

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build web types test dev icons clean

# The program, for this machine.
build: web
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o hermaeus ./cmd/hermaeus

# The frontend, into web/build, where the program embeds it from.
web: web/node_modules
	npm --prefix web run build

web/node_modules: web/package.json web/package-lock.json
	npm --prefix web ci --ignore-scripts
	@touch $@

# The icon files, from the drawings in assets/: the icon, and a tighter
# frame on the eye for the 16-pixel size, where the tentacles turn to mush.
# Needs ImageMagick; only run when a drawing changes, since the results are
# checked in.
icons:
	convert -background none -density 400 assets/hermaeus-icon-small.svg -resize 16x16 /tmp/hermaeus-16.png
	for n in 24 32 48 64 128 256; do convert -background none -density 400 assets/hermaeus-icon.svg -resize $${n}x$${n} /tmp/hermaeus-$$n.png; done
	convert /tmp/hermaeus-16.png /tmp/hermaeus-32.png /tmp/hermaeus-48.png web/static/favicon.ico
	convert /tmp/hermaeus-16.png /tmp/hermaeus-24.png /tmp/hermaeus-32.png /tmp/hermaeus-48.png /tmp/hermaeus-64.png /tmp/hermaeus-128.png /tmp/hermaeus-256.png cmd/hermaeus/hermaeus.ico
	convert -background none -density 400 assets/hermaeus-icon.svg -resize 180x180 web/static/apple-touch-icon.png
	cp assets/hermaeus-icon.svg web/static/favicon.svg
	rm -f /tmp/hermaeus-*.png

# TypeScript types for the API, from the Go types listed in tygo.yaml.
types:
	go tool tygo generate

# Go tests (which don't need the frontend built), then type-check the
# frontend, including that it still matches the generated API types.
test: web/node_modules types
	go test ./internal/...
	npm --prefix web run check

# Hermaeus plus Vite's dev server (http://localhost:5173), which reloads the
# page as the frontend changes. Hermaeus gets its own port, so a copy already
# running on the usual 8765 doesn't take its place. Both get the same token,
# since the page Vite serves can't carry it. Pass a SkyrimNet folder with DIR=…
# If either one stops (including Hermaeus failing to start), both do.
DEV_PORT ?= 8775

dev: web/node_modules
	@test -d web/build || npm --prefix web run build
	@export HERMAEUS_DEV_TOKEN=$$(openssl rand -hex 16) HERMAEUS_DEV_PORT=$(DEV_PORT); \
	trap 'kill 0' EXIT; \
	(go run ./cmd/hermaeus -no-browser -port $(DEV_PORT) $(if $(DIR),-dir "$(DIR)"); kill 0) & \
	npm --prefix web run dev

clean:
	rm -rf hermaeus web/build web/.svelte-kit
