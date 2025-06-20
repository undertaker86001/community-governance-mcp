.PHONY: build clean test

PLUGIN_NAME = community-governance-mcp
BUILD_TIME = $(shell date +%Y%m%d-%H%M%S)
VERSION = v1.0.0

build:
	tinygo build -o $(PLUGIN_NAME).wasm -scheduler=none -target=wasi -gc=custom -tags='custommalloc nottinygc_finalizer' ./main.go

clean:
	rm -f *.wasm

test:
	go test ./...

docker-build:
	docker run --rm -v $(PWD):/workspace -w /workspace tinygo/tinygo:0.30.0 make build

.DEFAULT_GOAL := build