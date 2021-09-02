#!/bin/bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

# map values to expected environment variables
export CLOUD_API_KEY=$IBM_CLOUD_API_KEY
# strip off preceding 'ibm:yp:' if present
export REGION=${IBM_CLOUD_REGION##*:}

# generate a random 32 character string for IBM Functions Action API key
export FN_WEB_SECURE_KEY=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

# login
ibmcloud login --apikey "${CLOUD_API_KEY}" -r "${REGION}" || { echo 'IBM Cloud CLI login failed!'; exit 1; }
ibmcloud target -g "${RESOURCE_GROUP}"

# Keep track of any failures
rtn=0

# Update ElasticSearch
echo "Configuring HRI Document Store (ElasticSearch)..."
./elastic.sh || { echo 'HRI Document Store (ElasticSearch) configuration failed!' ; rtn=1; }

if $SET_UP_APPID; then
  echo "Setting up AppId applications and roles..."
  ./appid.sh || { echo 'AppId setup failed!' ; rtn=1; }

  # JWT_AUDIENCE_ID was written to file -- read it, then delete it.
  export JWT_AUDIENCE_ID=$(cat JWT_AUDIENCE_ID)
  echo "JWT_AUDIENCE_ID set to $JWT_AUDIENCE_ID"
  rm JWT_AUDIENCE_ID
fi

# Deploy Mgmt-API
echo "Deploying HRI Management API..."
./deploy.sh || { echo 'HRI Management API Deployment failed!' ; rtn=1; }

# Write the smoke test results to the log
echo "smoketests.xml"
cat smoketests.xml

exit $rtn
