# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

# This creates an IBM Function's zipped actionloop executable for every *.go file in the base src/ directory.
# Each of these files should have a 'main' method and use common/actionloopmin to implement the actionloop protocol.
# Also '//go:build !tests' must be at the beginning of the file, so it is excluded from tests.
MAINS:= $(wildcard src/*.go)
BUILD:=build
BIN:=$(MAINS:src/%.go=$(BUILD)/%-bin.zip)
SRCS:=$(find src -name *.go)

# Multiple builds on the same input files is causing the Makefile's automatic rebuild only on dependency file modification to not work properly. Added clean to the default build, so the actions are always built.
bins: clean test format tidy $(BIN)

$(BIN): $(SRCS) src/go.mod src/go.sum $(BUILD)
	cd src; GOOS=linux GOACH=amd64 go build -o exec $(@:$(BUILD)/%-bin.zip=%.go)
	cd src; zip ../$@ -qr exec exec.env
	rm src/exec

# '-tags tests' is used to excluded the multiple main declarations from test builds
test:
	cd src; go test -coverprofile testCoverage.out ./... -v -tags tests

coverage:
	cd src; go tool cover -html=testCoverage.out

format:
	cd src; go fmt ./...

tidy: 
	cd src; go mod tidy

deps:
	cd src; go list -m all

clean:
	-rm -f $(BUILD)/*.zip 2>/dev/null

$(BUILD):
	mkdir $(BUILD)

# targets that don't produce physical files and get run every time
.PHONY: test clean tidy deps
