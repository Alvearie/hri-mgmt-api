# HRI Management API
The Alvearie Health Record Ingestion service: a common 'Deployment Ready Component' designed to serve as a “front door” for data for cloud-based solutions. See our [documentation](https://alvearie.io/HRI/) for more details.

This repo contains the code for the Management API of the HRI, which uses [IBM Functions](https://cloud.ibm.com/docs/openwhisk?topic=cloud-functions-getting-started) (Serverless built on [OpenWhisk](https://openwhisk.apache.org/)) with [Golang](https://golang.org/doc/). Basically, this repo defines an API and maps endpoints to Golang executables packaged into 'actions'. IBM Functions takes care of standing up an API Gateway, executing & scaling the actions, and transmitting data between them. [mgmt-api-manifest.yml](mgmt-api-manifest.yml) defines the actions, API, and the mapping between them. A separate OpenAPI specification is maintained in [Alvearie/hri-api-spec](https://github.com/Alvearie/hri-api-spec) for external user's reference. Please Note: Any changes to this (RESTful) Management API for the HRI requires changes in both the hri-api-spec repo and this hri-mgmt-api repo.

## Communication
* Please [join](https://alvearie.io/contributions/requestSlackAccess) our Slack channel for further questions: [#health-record-ingestion](https://alvearie.slack.com/archives/C01GM43LFJ6)
* Please see recent contributors or [maintainers](MAINTAINERS.md)

## Getting Started

### Prerequisites

* Golang 1.15 - you can use an official [distribution](https://golang.org/dl/) or a package manager like `homebrew` for mac
* Make - should come pre-installed on MacOS and Linux
* [GoMock latest](https://github.com/golang/mock) released version. Installation: 
    run `$ GO111MODULE=on go get github.com/golang/mock/mockgen@latest`. See [GoMock docs](https://github.com/golang/mock). 
* Golang IDE (Optional) - we use IntelliJ, but it requires a licensed version. VSCode is also supposed to be good and free.
* Ruby (Optional) - required for integration tests. See [testing](test/README.md) for more details.
* IBM Cloud CLI (Optional) - useful for local testing with IBM Functions. Installation [instructions](https://cloud.ibm.com/docs/cli?topic=cloud-cli-getting-started).

### Building

From the base directory, run `make`. This will download dependencies using Go Modules, run all the unit tests, and package up the code for IBM Functions in the `build` directory.

```
hri-mgmt-api$ make
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
Since this application must be deployed using IBM Functions in an IBM Cloud account, there isn't a way to launch and test the API & actions locally. So, we have set up GitHub actions to automatically deploy every branch in its own IBM Function's namespace in our IBM cloud account and run integration tests. They all share common Elastic Search and Event Streams instances. Once it's deployed, you can perform manual testing with your namespace. You can also use the IBM Functions UI or IBM Cloud CLI to modify the actions or API in your namespace. When the GitHub branch is deleted, the associated IBM Function's namespace is also automatically deleted. 

### Docker image build
Images are published on every `develop` branch build with the tag `<branch>-timestamp`.

## Code Overview

### IBM Function Actions - Golang Mains
For each API endpoint, there is a Golang executable packaged into an IBM Function's 'action' to service the requests. There are several `.go` files in the base `src/` directory, one for each action and no others, each of which defines `func main()`. If you're familiar with Golang, you might be asking how there can be multiple files with different definitions of `func main()`. The Makefile takes care of compiling each one into a separate executable, and each file includes a [Build Constraint](https://golang.org/pkg/go/build/#hdr-Build_Constraints) to exclude it from unit tests. This also means these files are not unit tested and thus are kept as small as possible. Each one sets up any required clients and then calls an implementation method in a sub package. They also use `common.actionloopmin.Main()` to implement the OpenWhisk [action loop protocol](https://github.com/apache/openwhisk-runtime-go/blob/main/docs/ACTION.md). 

The compiled binaries have to be named `exec` and put in a zip file. Additionally, a `exec.env` file has to be included, which contains the name of the docker container to use when running the action. All the zip files are written to the `build` directory when running `make`. 

### Packages

- tenants - code for all the `tenants` endpoints. Tenants are mainly indexes in Elastic Search.

- streams - code for all the `tenants/<tenantId>/streams` endpoints. Streams are mainly sets of topics in Kafka (Event Streams).

- batches - code for all the `tenants/<tenantId>/batches` endpoints.

- common - common code for various clients (i.e. Elastic Search & Kafka) and input/output helper methods.

- healthcheck - code for the `healthcheck` endpoint

### Unit Tests
Each unit test file follows the Golang conventions where it's named `*_test.go` (e.g. `get_by_id_test.go`) and is located in the same directory as the file it's testing (e.g. `get_by_id.go`). The team uses 'mock' code in files generated by [the GoMock framework](https://github.com/golang/mock) to support unit testing by mocking some collaborating component (function/struct) upon which the System Under Test (function/struct) depends. A good example test class that makes use of a mock object is `connector_test.go` (in `kafka` package) - it makes use of the generated code in `connector_mock.go`. Note that you may need to create a new go interface in order to use GoMock to generate the mock code file for the behavior you are trying to mock. This is because the framework requires [an interface](https://medium.com/rungo/interfaces-in-go-ab1601159b3a) in order to generate the mock code. [This article](https://medium.com/@duythhuynh/gomock-unit-testing-made-easy-b59a0e947ba7) might be helpful to understand GoMock. 

#### Test Coverage
The goal is to have 90% code coverage with unit tests. The build automatically prints out test coverage percentages and a coverage file at `src/testCoverage.out`. Additionally, you can view test coverage interactively in your browser by running `make coverage`.

### API Definition
The API that this repo implements is defined in [Alvearie/hri-api-spec](https://github.com/Alvearie/hri-api-spec) using OpenAPI 3.0. There are automated Dredd tests to make sure the implemented API meets the spec. If there are changes to the API, make them to the specification repo using a branch with the same name. Then the Dredd tests will run against the modified API specification. 

### Authentication & Authorization
All endpoints (except the health check) require an OAuth 2.0 JWT bearer access token per [RFC8693](https://tools.ietf.org/html/rfc8693) in the `Authorization` header field. The Tenant and Stream endpoints require IAM tokens, but the Batch endpoints require a token with HRI and Tenant scopes for authorization. The Batch token issuer is configurable via a bound parameter, and must be OIDC compliant because the code dynamically uses the OIDC defined well know endpoints to validate tokens. Integration and testing have already been completed with App ID, the standard IBM Cloud solution.

Batch JWT access token scopes:
- hri_data_integrator - Data Integrators can create, get, and change the status of batches, but only ones that they created.
- hri_consumer - Consumers can list and get Batches
- tenant_<tenantId> - provides access to this tenant's batches. This scope must use the prefix 'tenant_'. For example, if a data integrator tries to create a batch by making an HTTP POST call to `tenants/24/batches`, the token must contain scope `tenant_24`, where the `24` is the tenantId.
 
The scopes claim must contain one or more of the HRI roles ("hri_data_integrator", "hri_consumer") as well as the tenant id of the tenant being accessed.

## Contribution Guide
Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.
