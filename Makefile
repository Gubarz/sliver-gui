.PHONY: analyze analyze-go analyze-frontend build

GOCACHE ?= /tmp/sliver-gui-cache/go-build
XDG_CACHE_HOME ?= /tmp/sliver-gui-cache
CCACHE_DIR ?= /tmp/sliver-gui-cache/ccache
BUILD_OUTPUT ?= /tmp/sliver-gui-build

export GOCACHE
export XDG_CACHE_HOME
export CCACHE_DIR

analyze: analyze-go analyze-frontend

analyze-go:
	go test .
	go vet .
	staticcheck .
	deadcode -test .
	dupl -t 50 *.go

analyze-frontend:
	npm --prefix frontend run analyze

build:
	npm --prefix frontend run build
	go build -o $(BUILD_OUTPUT) .
