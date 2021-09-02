#!/usr/bin/env bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

2>&1

set -eo pipefail

echo "CLOUD_API_KEY: ****"
echo "REGION: $REGION"
echo "RESOURCE_GROUP: $RESOURCE_GROUP"
echo "NAMESPACE: $NAMESPACE"
echo "FN_WEB_SECURE_KEY: ****"
echo "ELASTIC_INSTANCE: $ELASTIC_INSTANCE"
echo "ELASTIC_SVC_ACCOUNT: $ELASTIC_SVC_ACCOUNT"
echo "KAFKA_INSTANCE: $KAFKA_INSTANCE"
echo "KAFKA_SVC_ACCOUNT: $KAFKA_SVC_ACCOUNT"
echo "OIDC_ISSUER: $OIDC_ISSUER"
echo "VALIDATION: $VALIDATION"

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

ibmcloud login --apikey "${CLOUD_API_KEY}" -r "${REGION}"
ibmcloud target -g "${RESOURCE_GROUP}"
ibmcloud fn property unset --namespace

# create namespace if it doesn't already exist
if ! ibmcloud fn property set --namespace "${NAMESPACE}"; then
    echo "creating IBM Functions namespace ${NAMESPACE}"
    ibmcloud fn namespace create "${NAMESPACE}"
    ibmcloud fn property set --namespace "${NAMESPACE}"
fi

# deploy hri-mgmt-api to IBM Functions
ibmcloud fn deploy --manifest mgmt-api-manifest.yml

echo "Building OpenWhisk Parameters"
# set config parameters, all of them have to be set in the same command
if [ -z "$VALIDATION" ] || [ "$VALIDATION" != true ]; then
    echo "Setting Validation to false"
    VALIDATION=false
else
    echo "Setting Validation to true"
    VALIDATION=true
fi

params="$(cat << EOF
{
  "issuer": "$OIDC_ISSUER",
  "validation": $VALIDATION,
  "jwtAudienceId": "$JWT_AUDIENCE_ID"
}
EOF
)"
echo $params

# save OpenWhisk parameters to temp file
paramFile=$(mktemp)
echo $params > $paramFile

echo "Created temp params file"

# set config parameters, all of them have to be set in the same command
ibmcloud fn package update hri_mgmt_api --param-file $paramFile

# cleanup temp file
rm $paramFile

# bind Elastic and Kafka service instances to hri-mgmt-api
ibmcloud fn service bind databases-for-elasticsearch hri_mgmt_api --instance "${ELASTIC_INSTANCE}" --keyname "${ELASTIC_SVC_ACCOUNT}"
ibmcloud fn service bind messagehub hri_mgmt_api --instance "${KAFKA_INSTANCE}" --keyname "${KAFKA_SVC_ACCOUNT}"

./run-smoketests.sh
