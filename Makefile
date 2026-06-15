VERSION ?= dev
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: test bench build release trellis-lint

test:
	go test ./...

bench:
	go test -bench=. ./...

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags "-X github.com/norlinga/loupe/internal/version.Version=$(VERSION)" -o bin/loupe .

release:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "-s -w -X github.com/norlinga/loupe/internal/version.Version=$(VERSION)" -o dist/loupe-$(GOOS)-$(GOARCH) .

trellis-lint:
	trellis lint main.go.trellis internal/schema/schema.go.trellis internal/observe/observe.go.trellis internal/context/context.go.trellis internal/notes/notes.go.trellis internal/version/version.go.trellis internal/cli/cli.go.trellis
