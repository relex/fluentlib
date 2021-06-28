AUTO_BUILD_VERSION ?= dev
export LINT_EXHAUSTIVESTRUCT=Y

build: BUILD/fluentlibtool

include ${GOPATH}/opt/gotils/Common.mk

BUILD/fluentlibtool: Makefile go.mod $(SOURCES_NONTEST)
	GO_LDFLAGS="-X main.version=$(AUTO_BUILD_VERSION)" gotils-build.sh -o $@
