# HRI Management API
The IBM Watson Health, Health Record Ingestion service is an open source project designed to serve as a “front door”, receiving health data for cloud-based solutions. See our [documentation](https://alvearie.github.io/HRI/) for more details.

This repo contains the code for the Management API of the HRI, which uses [IBM Functions](https://cloud.ibm.com/docs/openwhisk?topic=cloud-functions-getting-started) (Serverless built on [OpenWhisk](https://openwhisk.apache.org/)) with [Golang](https://golang.org/doc/). This repo defines an API and maps endpoints to Golang executables packaged into 'actions'. IBM Functions takes care of standing up an API Gateway, executing & scaling the actions, and transmitting data between the gateway and action endpoints. [mgmt-api-manifest.yml](mgmt-api-manifest.yml) defines the actions, API, and the mapping between them. A separate OpenAPI specification is maintained in [Alvearie/hri-api-spec](https://github.com/Alvearie/hri-api-spec) for external user's reference. Please Note: Any changes to this (RESTful) Management API for the HRI requires changes in both the api-spec repo and this mgmt-api repo.

This is an initial publish of the code, which is being transitioned to open source, but is not yet completed. To use this code, you will need to update the CI/CD for your environment. There is a TravisCI build and integration tests, but all the references to the CD Process that deployed to internal IBM Cloud environments were removed. At some future date, a more comprehensive CI/CD pipeline will be published, as part of Watson Health's continuing Open Source initiative.

## Communication
* Please TBD
* Please see [MAINTAINERS.md](MAINTAINERS.md)

## Getting Started

### Prerequisites

* Golang 1.13 - you can use an official [distribution](https://golang.org/dl/) or a package manager like `homebrew` for mac
* Make - should come pre-installed on MacOS and Linux
* [GoMock latest](https://github.com/golang/mock) released version. Installation: 
    run `$ GO111MODULE=on go get github.com/golang/mock/mockgen@latest`. See [GoMock docs](https://github.com/golang/mock). 
* Golang IDE (Optional) - we use IntelliJ, but it requires a licensed version. VSCode is also supposed to be good and free.
* Ruby (Optional) - required for integration tests. See [testing](test/README.md) for more details.
* IBM Cloud CLI (Optional) - useful for local testing with IBM Functions. Installation [instructions](https://cloud.ibm.com/docs/cli?topic=cloud-cli-getting-started).

### Building

From the base directory, run `make`. This will download dependencies using Go Modules, run all the unit tests, and package up the code for IBM Functions in the `build` directory.

```
mgmt-api$ make
rm -f build/*.zip 2>/dev/null
cd src; go test ./... -v -tags tests
=== RUN   TestEsDocToBatch
=== RUN   TestEsDocToBatch/example1
--- PASS: TestEsDocToBatch (0.00s)
    --- PASS: TestEsDocToBatch/example1 (0.00s)
...
PASS
ok  	github.com/Alvearie/hri-mgmt-api/healthcheck	2.802s
cd src; GOOS=linux GOACH=amd64 go build -o exec batches_create.go
cd src; zip ../build/batches_create-bin.zip -qr exec
rm src/exec
...
cd src; GOOS=linux GOACH=amd64 go build -o exec healthcheck_get.go
cd src; zip ../build/healthcheck_get-bin.zip -qr exec
rm src/exec
```
## CI/CD
Since this application must be deployed using IBM Functions in an IBM Cloud account, there isn't a way to launch and test the API & actions locally. So, we have an automated TravisCI build that automatically deploys every branch in its own IBM Functions' namespace in the IBM Cloud and runs integration tests. They all share a common ElasticSearch and Event Streams instance. Once it's deployed, you can perform manual testing with your namespace. You can also use the IBM Functions UI or IBM Cloud CLI to modify the actions or API in your namespace.

If you're working off a fork, the secure `.travis.yml` environment variables will not work, because they are repository specific. We will work with you to get your fork deployed and tested. 

## Code Overview

### IBM Function Actions - Golang Mains
For each API endpoint, there is a Golang executable packaged into an IBM Function's 'action' to service the requests. There are several `.go` files in the base `src/` directory, one for each action and no others, each of which defines `func main()`. If you're familiar with Golang, you might be asking how there can be multiple files with different definitions of `func main()`. The Makefile takes care of compiling each one into a separate executable, and each file includes a [Build Constraint](https://golang.org/pkg/go/build/#hdr-Build_Constraints) to exclude it from unit tests. This also means these files are not unit tested and thus are kept as small as possible. Each one sets up any required clients and then calls an implementation method in a sub package. They also use `common.actionloopmin.Main()` to implement the OpenWhisk [action loop protocol](https://github.com/apache/openwhisk-runtime-go/blob/master/docs/ACTION.md).

### Packages
There are three packages:
- batches - code for all the `tenants/tenantId/batches` endpoints. In general, each endpoint has an implementation method that each `func main()` above calls.

- common - common code for various clients (i.e. ElasticSearch & Kafka) and input/output helper methods.

- healthcheck - code for the `healthcheck` endpoint

### Unit Tests
Each unit test file follows the Golang conventions where it's named `*_test.go` (e.g. `get_by_id_test.go`) and is located in the same directory as the file it's testing (e.g. `get_by_id.go`). The team uses 'mock' code in files generated by [the GoMock framework](https://github.com/golang/mock) to support unit testing by mocking some collaborating component (function/struct) upon which the System Under Test (function/struct) depends. A good example test class that makes use of a mock object is `connector_test.go` (in `kafka` package) - it makes use of the generated code in `connector_mock.go`. Note that you may need to create a new go interface in order to use GoMock to generate the mock code file for the behavior you are trying to mock. This is because the framework requires [an interface](https://medium.com/rungo/interfaces-in-go-ab1601159b3a) in order to generate the mock code. [This article](https://medium.com/@duythhuynh/gomock-unit-testing-made-easy-b59a0e947ba7) might be helpful to understand GoMock. 

#### Test Coverage
The goal is to have 90% code coverage with unit tests. The build automatically prints out test coverage percentages and a coverage file at `src/testCoverage.out`. Additionally, you can view test coverage interactively in your browser by running `go tool cover -html=testCoverage.out` from the `src/` directory.

### API Definition
The API that this repo implements is defined in [Alvearie/hri-api-spec](https://github.com/Alvearie/hri-api-spec) using OpenAPI 3.0. There are automated Dredd tests to make sure the implemented API meets the spec. For any changes that affect the API itself (which naturally, result in changes also here in the mgmt-api Repo), please use branches with the same name for your corresponding changes in both repositories.

## Contribution Guide
Since we have not completely moved our development into the open yet, external contributions are limited. If you would like to make contributions, please create an issue detailing the change. We will work with you to get it merged in. 
