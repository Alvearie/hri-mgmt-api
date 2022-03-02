# This creates a binary executable file, src/hri, which creates a web server.
# Also '// +build !tests' must be at the beginning of the file, so it is excluded from tests.
# While the other functions are still in an unconverted state, they will also contain an // +build ignore which prevents the compiler from complaining about there being several definitions of func main().
HRI:=./src/hri
SRCS:=$(find src -name *.go)
TEST:=./src/testCoverage.out

$(HRI): clean format tidy $(TEST) src/go.mod src/go.sum $(SRCS)
	cd src; go build -o hri

test: clean $(TEST)

# '-tags tests' is used to excluded the multiple main declarations from test builds
$(TEST): $(SRCS) src/go.mod src/go.sum
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
	-rm ./src/hri
	-rm ./src/testCoverage.out

# targets that don't produce physical files and get run every time
.PHONY: coverage clean format tidy deps
