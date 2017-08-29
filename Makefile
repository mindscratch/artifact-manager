NAME=artifact-manager
IMAGE_NAME=$(NAME)
VERSION=0.0.3
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell dirname $(mkfile_path))
GO_TMPDIR=$(HOME)/tmp

GO ?= go

build: ## build the application
	@CGO_ENABLE=0 $(GO) build -a -ldflags '-s' -o $(NAME) .

test:
	@mkdir -p $(GO_TMPDIR)
	@TMPDIR=$(GO_TMPDIR) $(GO) test $$($(GO) list ./... | grep -v /vendor/)

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort
