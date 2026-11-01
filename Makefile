# Building Hermaeus needs Go and Node. Running it needs neither: the
# program it builds has the frontend embedded.

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOBUILD = CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build windows release web notices types test dev icons clean

# The program, for this machine.
build: web
	$(GOBUILD) -o hermaeus ./cmd/hermaeus

# The program for 64-bit Windows, as Hermaeus.exe, from any OS. Its icon and
# file details go into a .syso file that go build links into Windows builds.
windows: web
	go tool go-winres simply --arch amd64 --manifest cli --icon cmd/hermaeus/hermaeus.ico \
		--product-name Hermaeus --file-description "Hermaeus, an unofficial dashboard for SkyrimNet" \
		--original-filename Hermaeus.exe --out cmd/hermaeus/rsrc
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o Hermaeus.exe ./cmd/hermaeus

# A release's downloads, into dist/: a zip for Windows, a zip for macOS (one
# program for both Apple silicon and Intel), and a tarball for Linux, each
# with LICENSE and THIRD_PARTY_NOTICES.txt, plus their SHA256SUMS. Their
# names carry no version, so the site can link to the latest release's.
# Needs macOS, for lipo and codesign; the Release workflow runs it when a
# version tag is pushed.
release: windows
	rm -rf dist && mkdir -p dist/windows dist/macos dist/linux
	cp Hermaeus.exe dist/windows/
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o dist/macos/hermaeus-arm64 ./cmd/hermaeus
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o dist/macos/hermaeus-amd64 ./cmd/hermaeus
	lipo -create -output dist/macos/hermaeus dist/macos/hermaeus-arm64 dist/macos/hermaeus-amd64
	rm dist/macos/hermaeus-arm64 dist/macos/hermaeus-amd64
	codesign --force --sign - dist/macos/hermaeus
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/linux/hermaeus ./cmd/hermaeus
	for os in windows macos linux; do cp LICENSE dist/$$os/LICENSE.txt; cp THIRD_PARTY_NOTICES.txt dist/$$os/; done
	cd dist/windows && zip -qX ../Hermaeus-windows-x64.zip Hermaeus.exe LICENSE.txt THIRD_PARTY_NOTICES.txt
	cd dist/macos && zip -qX ../Hermaeus-macos.zip hermaeus LICENSE.txt THIRD_PARTY_NOTICES.txt
	COPYFILE_DISABLE=1 tar -czf dist/Hermaeus-linux-x64.tar.gz -C dist/linux hermaeus LICENSE.txt THIRD_PARTY_NOTICES.txt
	cd dist && shasum -a 256 Hermaeus-* > SHA256SUMS

# THIRD_PARTY_NOTICES.txt: the license of everything built into the program,
# from its Go modules and the npm packages in its pages, which the frontend's
# build records. Run it after changing dependencies; releases ship it beside
# LICENSE.
notices: web
	go run ./cmd/notices

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
	rm -rf hermaeus Hermaeus.exe dist cmd/hermaeus/*.syso web/build web/.svelte-kit
