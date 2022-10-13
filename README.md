# HRI Management API
The Alvearie Health Record Ingestion service: a common 'Deployment Ready Component' designed to serve as a “front door” for data for cloud-based solutions. See our [documentation](https://alvearie.io/HRI/) for more details.

This repo contains the code for the Management API of the HRI, which is written in Golang using the [Echo](https://echo.labstack.com/) web framework. A separate OpenAPI specification is maintained in [Alvearie/hri-api-spec](https://github.com/Alvearie/hri-api-spec) for external user's reference. Please Note: Any changes to this (RESTful) Management API for the HRI requires changes in both the hri-api-spec repo and this hri-mgmt-api repo.

This version is compatible with HRI `v3.1`.

## Communication
* Please [join](https://alvearie.io/contributions/requestSlackAccess) our Slack channel for further questions: [#health-record-ingestion](https://alvearie.slack.com/archives/C01GM43LFJ6)
* Please see recent contributors or [maintainers](MAINTAINERS.md)

## Getting Started

### Prerequisites

* Golang 1.17 - you can use an official [distribution](https://golang.org/dl/) or a package manager like `homebrew` for mac
* Make - should come pre-installed on MacOS and Linux
* [GoMock latest](https://github.com/golang/mock) released version. Installation: 
    run `$ go get github.com/golang/mock/mockgen@latest`. See [GoMock docs](https://github.com/golang/mock). 
* Golang IDE (Optional) - we use IntelliJ, but it requires a licensed version. VSCode is also supposed to be good and free.
* Ruby (Optional) - required for integration tests. See [testing](test/README.md) for more details.

### Building

From the base directory, run `make`. This will download dependencies using Go Modules, run all the unit tests, and produce an executable at `src/hri`.

```
hri-mgmt-api$ make
rm ./src/hri
rm ./src/testCoverage.out
cd src; go fmt ./...
cd src; go mod tidy
cd src; go test -coverprofile testCoverage.out ./... -v -tags tests
?       github.com/Alvearie/hri-mgmt-api    [no test files]
=== RUN   TestEsDocToBatch
=== RUN   TestEsDocToBatch/example1
--- PASS: TestEsDocToBatch (0.00s)
    --- PASS: TestEsDocToBatch/example1 (0.00s)...
PASS
coverage: 100.0% of statements
ok  	github.com/Alvearie/hri-mgmt-api/tenants	0.316s	coverage: 100.0% of statements
cd src; GOOS=linux GOACH=amd64 go build
```

### Troubleshooting
If you encounter this error:
```
rdkafka#producer-1| [thrd:sasl_ssl://...]: sasl_ssl://.../bootstrap: SSL handshake failed: error:14090086:SSL routines:ssl3_get_server_certificate:certificate verify failed: broker certificate could not be verified, verify that ssl.ca.location is correctly configured or root CA certificates are installed (brew install openssl)
```
There is a problem with the default root CA. See the Confluent [documentation](https://docs.confluent.io/platform/current/tutorials/examples/clients/docs/go.html#configure-ssl-trust-store) for instructions on how to fix it.

## CI/CD
This application can be run locally, but almost all the endpoints require Elasticsearch, Kafka, and an OIDC server. GitHub actions builds and runs integration tests using a common Elastic Search and Event Streams instance. You can perform local manual testing using these resources. See [test/README.md](test/README.md) for more details.

### Static Code Analysis
In addition, SonarCloud [analysis](https://sonarcloud.io/dashboard?id=Alvearie_hri-mgmt-api) is performed on all pull requests and on long-lived branches: `main`, `develop`, and `support-*`. Several IDE's, including IntelliJ, have SonarLint plugins for dynamic analysis as you code.

### Dependency Vulnerabilities
Dependencies are also checked for vulnerabilities when a pull request is created. If any are found, a comment will be added to the pull request with a link to the GitHub action, where you can view the logs for the details. If the vulnerabilities are caused by your changes or easy to fix, implement the changes on your branch. If not, create a ticket. To run the check again, submit a review for the pull request with `/pr_checks` in the message. 

### Releases
Releases are created by creating GitHub tags, which trigger a build that packages everything into a Docker image. See [docker/README.md](docker/README.md) for more details.

### Docker image build
Images are published on every `develop` build with the tag `develop-timestamp`. See [Overall strategy](https://github.com/Alvearie/HRI/wiki/Overall-Project-Branching,-Test,-and-Release-Strategy) for more details.

## Code Overview

### Serve.go
`src/serve.go` defines the main method where execution begins. It reads the config, creates the Echo server, creates and registers handlers, and then starts the server.

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

In addition to the API spec, an `/alive` endpoint was added to support Kubernetes readiness and liveness probes. This endpoint returns `yes` with a 200 response code when the Echo web server is up and running.

### Authentication & Authorization
All endpoints (except the health check) require an OAuth 2.0 JWT bearer access token per [RFC8693](https://tools.ietf.org/html/rfc8693) in the `Authorization` header field. The Tenant and Stream endpoints require IAM tokens, but the Batch endpoints require a token with HRI and Tenant scopes for authorization. The Batch token issuer is configurable via a bound parameter, and must be OIDC compliant because the code dynamically uses the OIDC defined well know endpoints to validate tokens. Integration and testing have already been completed with [App ID](https://cloud.ibm.com/docs/appid), the standard IBM Cloud solution.

Batch JWT access token scopes:
- hri_data_integrator - Data Integrators can create, get, and call 'sendComplete' and 'terminate' endpoints for batches, but only ones that they created.
- hri_consumer - Consumers can list and get batches.
- hri_internal - For internal processing, can call batch 'processingComplete' and 'fail' endpoints.
- tenant_<tenantId> - provides access to this tenant's batches. This scope must use the prefix 'tenant_'. For example, if a data integrator tries to create a batch by making an HTTP POST call to `tenants/24/batches`, the token must contain scope `tenant_24`, where the `24` is the tenantId.
 
The scopes claim must contain one or more of the HRI roles ("hri_data_integrator", "hri_consumer", "hri_internal") as well as the tenant id of the tenant being accessed.

## Contribution Guide
Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.