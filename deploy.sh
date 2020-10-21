#!/usr/bin/env bash
set -eo pipefail

# check for command-line options
SKIP_SECURE_API=false
POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"
    case $key in
    -s|--skipSecureApi)
        SKIP_SECURE_API=true
        shift
        ;;
    *)
        POSITIONAL+=("$1")
        shift
        ;;
    esac
done
set -- "${POSITONAL[@]}" # restore unprocessed positional params

# prompt for parameters if not already provided as environment variables
if [ -z "$CLOUD_API_KEY" ]
then
    read -p "CLOUD_API_KEY not set. Enter IBM Cloud API Key: " CLOUD_API_KEY
fi

if [ -z "$REGION" ]
then
    read -p "REGION not set. Enter target Region (i.e. us-south): " REGION
fi

if [ -z "$RESOURCE_GROUP" ]
then
    read -p "RESOURCE_GROUP not set. Enter target Resource Group (i.e. CDT_Payer_ASB_RG): " RESOURCE_GROUP
fi

if [ -z "$NAMESPACE" ]
then
    read -p "NAMESPACE not set. Enter target namespace: " NAMESPACE
fi

if [ -z "$FN_WEB_SECURE_KEY" ]
then
    read -p "FN_WEB_SECURE_KEY not set. Enter IBM Function Web API Key: " FN_WEB_SECURE_KEY
fi

if [ -z "$ELASTIC_INSTANCE" ]
then
    read -p "ELASTIC_INSTANCE not set. Enter Elastic instance name (i.e. HRI-DocumentStore): " ELASTIC_INSTANCE
fi

if [ -z "$ELASTIC_SVC_ACCOUNT" ]
then
    read -p "ELASTIC_SVC_ACCOUNT not set. Enter Elastic service account: " ELASTIC_SVC_ACCOUNT
fi

if [ -z "$KAFKA_INSTANCE" ]
then
    read -p "KAFKA_INSTANCE not set. Enter Kafka instance name (i.e. HRI-Event Streams): " KAFKA_INSTANCE
fi

if [ -z "$KAFKA_SVC_ACCOUNT" ]
then
    read -p "KAFKA_SVC_ACCOUNT not set. Enter Kafka service account: " KAFKA_SVC_ACCOUNT
fi

echo "CLOUD_API_KEY: ****"
echo "REGION: $REGION"
echo "RESOURCE_GROUP: $RESOURCE_GROUP"
echo "NAMESPACE: $NAMESPACE"
echo "FN_WEB_SECURE_KEY: ****"
echo "ELASTIC_INSTANCE: $ELASTIC_INSTANCE"
echo "ELASTIC_SVC_ACCOUNT: $ELASTIC_SVC_ACCOUNT"
echo "KAFKA_INSTANCE: $KAFKA_INSTANCE"
echo "KAFKA_SVC_ACCOUNT: $KAFKA_SVC_ACCOUNT"

# determine if IBM Cloud CLI is already installed
set +e > /dev/null 2>&1
ibmcloud --version > /dev/null 2>&1
res=$?
set -e > /dev/null 2>&1

if test "$res" != "0"; then
    # install IBM Cloud CLI
    curl -sL https://ibm.biz/idt-installer | bash
fi

# determine if IBM Functions CLI is already installed
set +e > /dev/null 2>&1
ibmcloud fn > /dev/null 2>&1
res=$?
set -e > /dev/null 2>&1

if test "$res" != "0"; then
    # install IBM Functions CLI
    ibmcloud plugin install cloud-functions
fi

ibmcloud login --apikey $CLOUD_API_KEY -r $REGION
ibmcloud target -g $RESOURCE_GROUP 
ibmcloud fn property unset --namespace

# create namespace if it doesn't already exist
if ! ibmcloud fn property set --namespace ${NAMESPACE}; then
    echo "creating IBM Functions namespace ${NAMESPACE}"
    ibmcloud fn namespace create $NAMESPACE
    ibmcloud fn property set --namespace $NAMESPACE
fi

# deploy HRI mgmt-api to IBM Functions
ibmcloud fn deploy --manifest mgmt-api-manifest.yml

# redefine IBM Functions API to properly authenticate with web actions
# (only necessary because of the following openwhisk bug: )
./updateApi.sh

# optionally require calls to the API endpoint to be secured by an API key
# (there is currently no way to automate the generation of API keys)
if [ "$SKIP_SECURE_API" != true ]; then
    ./secureApi.sh;
fi

# bind Elastic and Kafka service instances to HRI mgmt-api
ibmcloud fn service bind databases-for-elasticsearch hri_mgmt_api --instance "$ELASTIC_INSTANCE" --keyname $ELASTIC_SVC_ACCOUNT
ibmcloud fn service bind messagehub hri_mgmt_api --instance "$KAFKA_INSTANCE" --keyname $KAFKA_SVC_ACCOUNT

./run-smoketests.sh
