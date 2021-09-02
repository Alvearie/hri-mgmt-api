#!/usr/bin/env bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0
#
# This sript will set the 'validation' property on the Management API to 'true' or 'false'. This property changes the behavior of the API based on whether there is Flink Validation. The main use for this script is to support integration testing of both behaviors.
set -eo pipefail

validation=$1

if [ -z "$validation" ] || [ "$validation" != "true" ] && [ "$validation" != "false" ];
then
    echo "Missing validation parameter or invalid value."
    echo "Usage: ./setValidation.sh ( true | false )"
    exit 1
fi

if [ -z "$ELASTIC_INSTANCE" ]
then
    echo "Please set ELASTIC_INSTANCE environment variable"
    exit 1
fi

if [ -z "$ELASTIC_SVC_ACCOUNT" ]
then
    echo "Please set ELASTIC_SVC_ACCOUNT environment variable"
    exit 1
fi

if [ -z "$KAFKA_INSTANCE" ]
then
    echo "Please set KAFKA_INSTANCE environment variable"
    exit 1
fi

if [ -z "$KAFKA_SVC_ACCOUNT" ]
then
    echo "Please set KAFKA_SVC_ACCOUNT environment variable"
    exit 1
fi

if [ -z "$OIDC_ISSUER" ]
then
    echo "Please set OIDC_ISSUER environment variable"
    exit 1
fi

if [ -z "$JWT_AUDIENCE_ID" ]
then
    read -p "JWT_AUDIENCE_ID not set. Enter the Audience ID: " JWT_AUDIENCE_ID
fi

echo "Setting validation to $validation"

#NOTE: updating a package with paramters, will overwrite all the existing parameters including bound credentials. So if more parameters are ever added, they will need to be included in this script. And the service bind commands always need to be rerun at the end.

echo "Building OpenWhisk Parameters"
params="$(cat << EOF
{
  "issuer": "$OIDC_ISSUER",
  "validation": $validation,
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

rm $paramFile

# bind Elastic and Kafka service instances to hri-mgmt-api
ibmcloud fn service bind databases-for-elasticsearch hri_mgmt_api --instance "${ELASTIC_INSTANCE}" --keyname "${ELASTIC_SVC_ACCOUNT}"
ibmcloud fn service bind messagehub hri_mgmt_api --instance "${KAFKA_INSTANCE}" --keyname "${KAFKA_SVC_ACCOUNT}"

