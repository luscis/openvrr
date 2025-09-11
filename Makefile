SHELL := /bin/bash

ARCH ?=amd64
SD = $(shell pwd)
BD = $(SD)/build

env:
	mkdir -p $(BD)
	go version
	gofmt -w -s ./pkg ./cmd

vendor:
	go clean -modcache
	go mod tidy
	go mod vendor -v

cmd: env
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -o $(BD)/openvrr ./cmd/main.go