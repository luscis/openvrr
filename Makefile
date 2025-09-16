SHELL := /bin/bash

ARCH ?=amd64
SD = $(shell pwd)
BD = $(SD)/build

.PHONY: env vendor

auto: vrr cli

env:
	mkdir -p $(BD)
	go version
	gofmt -w -s ./pkg ./cmd

vendor:
	go clean -modcache
	go mod tidy
	go mod vendor -v

cli: env
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -o $(BD)/vrrcli ./cmd/cli/main.go

vrr: env
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -o $(BD)/openvrr ./cmd/router/main.go