# Docker Deploy
Deployment of the HRI Management API and configuring the ElasticSearch Document Store is packaged into a docker image. The image contains several scripts, all the compiled Go code, and Elastic index templates. It also runs smoke tests and logs the results.

There are several environment variables that must be set in the container.

|  Name     | Description         |
|-----------|---------------------|
| IBM_CLOUD_API_KEY   | The API key for IBM Cloud |
| IBM_CLOUD_REGION    | Target IBM Cloud Region, e.g. 'ibm:yp:us-south' |
| RESOURCE_GROUP      | Target IBM Cloud Resource Group |
| NAMESPACE           | Target IBM Function namespace |
| ELASTIC_INSTANCE    | Name of Elasticsearch instance |
| ELASTIC_SVC_ACCOUNT | Name of Elasticsearch service ID |
| KAFKA_INSTANCE      | Name of Event Streams (Kafka) instance |
| KAFKA_SVC_ACCOUNT   | Name of Event Streams (Kafka) service ID |
| VALIDATION          | Whether to deploy the Management API with Validation, e.g. 'true', 'false' |
| OIDC_ISSUER         | The base URL of the OIDC issuer to use for OAuth authentication (e.g. `https://us-south.appid.cloud.ibm.com/oauth/v4/<tenantId>`)               |
| APPID_PREFIX        | (Optional) Prefix string to append to the AppId applications and roles created during deployment                                                |
| SET_UP_APPID        | (Optional) defaults to true. Set to false if you do not want the App ID set-up enabled. |

## Implementation Details

The image entrypoint is `run.sh`, which:
 1. sets some environment variables
 1. logs into the IBM Cloud CLI
 1. calls `elastic.sh`
 1. calls `appid.sh`
 1. calls `deploy.sh`

`elastic.sh` turns off automatic index creation and sets the default template for batch indexes. These are idempotent actions, so they can be executed multiple times.

`appid.sh` creates HRI and HRI Internal applications and HRI Internal, HRI Consumer, and HRI Data Integrator roles in AppId.

`deploy.sh` deploys the Management API to IBM Functions and runs smoke tests (by calling the health check endpoint).

## Building
To build the docker image locally, build the code and run the following command from the base project directory:
```shell script
docker build ./ -f docker/Dockerfile
```

## Testing
You can test locally by running the docker container with all the required environment variables. A `template.env` file is included that you can copy and set the secure parameters. Be careful not to commit and secure parameters into GitHub.

```shell script
docker run --env-file docker/test.env --rm <image>
```

To investigate issues, you may want to run it interactively with a bash prompt. Just add `-it --entrypoint bash`.
```shell script
docker run --env-file docker/test.env --rm -it --entrypoint bash <image>
```
