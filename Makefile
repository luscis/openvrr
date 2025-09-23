SHELL := /bin/bash

VER = $(shell ./docker/script/version.sh)
ARCH ?=amd64
SD = $(shell pwd)
BD = $(SD)/build

.PHONY: env vendor docker

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
	GOOS=linux GOARCH=$(ARCH) go build -mod=vendor -o $(BD)/openvrr ./cmd/vrr/main.go

docker:
	cp -rf $(SD)/docker/Dockerfile $(BD)
	cp -rf $(SD)/docker/script/*.sh $(BD)
	cd $(BD) && sudo docker build -t luscis/openvrr:$(VER).$(ARCH) \
	--build-arg http_proxy="${http_proxy}" --build-arg https_proxy="${https_proxy}" \
	--file Dockerfile .
