AUTO_BUILD_VERSION ?= dev
GOPATH := $(shell go env GOPATH)
export LINT_EXHAUSTIVESTRUCT=Y

build: BUILD/fluentlibtool

include ${GOPATH}/opt/gotils/Common.mk

BUILD/fluentlibtool: Makefile go.mod $(SOURCES_NONTEST)
	GO_LDFLAGS="-X main.version=$(AUTO_BUILD_VERSION)" gotils-build.sh -o $@

.PHONY: test-gen
test-gen: BUILD/fluentlibtool
	LOG_LEVEL=$${LOG_LEVEL:-warn} LOG_COLOR=$${LOG_COLOR:-Y} go test -timeout $${TEST_TIMEOUT:-10s} -v ./... -args gen
