### Makefile for failpoint-ctl

LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.ReleaseVersion=$(shell git describe --tags --dirty="-dev")"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "github.com/pingcap/failpoint/failpoint-ctl/version.GoVersion=$(shell go version)"

FAILPOINT_CTL_BIN := bin/failpoint-ctl

path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(GOPATH)))
export PATH := $(path_to_add):$(PATH)

GO        := go
GOBUILD   := GO111MODULE=on CGO_ENABLED=0 $(GO) build
GOTEST    := GO111MODULE=on CGO_ENABLED=1 $(GO) test -p 3

ARCH      := "`uname -s`"
LINUX     := "Linux"
MAC       := "Darwin"

RACE_FLAG =
ifeq ("$(WITH_RACE)", "1")
	RACE_FLAG = -race
	GOBUILD   = GOPATH=$(GOPATH) CGO_ENABLED=1 $(GO) build
endif

.PHONY: build checksuccess

default: build checksuccess

build:
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS)' -o $(FAILPOINT_CTL_BIN) failpoint-ctl/main.go

checksuccess:
	@if [ -f $(FAILPOINT_CTL_BIN) ]; \
	then \
		echo "failpoint-ctl build successfully :-) !" ; \
	fi