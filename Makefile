GOPATH := $(shell go env GOPATH)
OUTPUT ?= BUILD/fluentlibtool

include ${GOPATH}/opt/gotils/Common.mk

.PHONY: test-gen
test-gen: BUILD/fluentlibtool
	LOG_LEVEL=$${LOG_LEVEL:-warn} LOG_COLOR=$${LOG_COLOR:-Y} go test -timeout $${TEST_TIMEOUT:-10s} -v ./... -args gen
