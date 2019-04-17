### Makefile for failpoint-ctl

LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.releaseVersion=$(shell git describe --tags --dirty="-dev" --always)"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.buildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.gitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.gitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.goVersion=$(shell go version)"

FAILPOINT_CTL_BIN := bin/failpoint-ctl

path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(GOPATH)))
export PATH := $(path_to_add):$(PATH)

GO        := GO111MODULE=on go
GOBUILD   := GO111MODULE=on CGO_ENABLED=0 $(GO) build
GOTEST    := GO111MODULE=on GO_FAILPOINTS="failpoint-env1=return(10);failpoint-env2=return(true)" GO_FAILPOINTS_HTTP=":23389" CGO_ENABLED=1 $(GO) test -p 4
OVERALLS  := CGO_ENABLED=1 GO111MODULE=on GO_FAILPOINTS="failpoint-env1=return(10);failpoint-env2=return(true)" GO_FAILPOINTS_HTTP=":23389" overalls

ARCH      := "`uname -s`"
LINUX     := "Linux"
MAC       := "Darwin"

RACE_FLAG =
ifeq ("$(WITH_RACE)", "1")
	RACE_FLAG = -race
	GOBUILD   = GOPATH=$(GOPATH) CGO_ENABLED=1 $(GO) build
endif

.PHONY: build checksuccess test cover upload-cover gotest

default: build checksuccess

build:
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS)' -o $(FAILPOINT_CTL_BIN) failpoint-ctl/main.go

checksuccess:
	@if [ -f $(FAILPOINT_CTL_BIN) ]; \
	then \
		echo "failpoint-ctl build successfully :-) !" ; \
	fi

test: gotest 

gotest:
	$(GOTEST) -v ./...

cover:
	$(GO) get github.com/go-playground/overalls
	$(OVERALLS) -project=github.com/pingcap/failpoint \
	    -covermode=count \
			-ignore='.git,vendor,LICENSES' \
			-concurrency=4
	
upload-cover:	SHELL:=/bin/bash
upload-cover:
	mv overalls.coverprofile coverage.txt
	bash <(curl -s https://codecov.io/bash)

