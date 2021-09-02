# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

#!/bin/bash

# Exit on errors
set -e

# Replace oauth with management in the OIDC_ISSUER url
issuer="${OIDC_ISSUER/oauth/management}"
echo "issuer:$issuer"

# Get IAM Token
# Note, in this command and many below, the response is gathered and then sent to jq via echo (rather than piping directly) because if you pipe the response
# directly to jq, the -f flag to fail if the curl command fails will not terminate the script properly.
echo "Requesting IAM token"
response=$(curl -X POST -sS 'https://iam.cloud.ibm.com/identity/token' -d "grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=${CLOUD_API_KEY}")
iamToken=$(echo $response | jq -r '.access_token // "NO_TOKEN"')
if [ $iamToken = "NO_TOKEN" ]; then
  echo "the_curl_response: $response"
  echo "Error getting IAM Token! Exiting!"
  exit 1
fi  

# Create application
echo
echo "Creating HRI provider application"
# Do not fail script if this call fails. We need to check if it failed because of a CONFLICT, in which case the script will exit 0.
hriApplicationName="${APPID_PREFIX}HRI Management API"
response=$(curl -X POST -sS "${issuer}/applications" -H "Authorization: Bearer ${iamToken}" -H 'Content-Type: application/json' -d @- << EOF
{
"name": "${hriApplicationName}",
"type": "regularwebapp"
}
EOF
)

# Get the application ID if the call was successful. If it was unsuccessful, check the returned code.
# If it failed because of a CONFLICT, that just means AppId was already configured (likely in a previous
# deployment attempt) and we will exit 0. If any other code, the script fails.
hriApplicationId=$(echo $response | jq -r '.clientId // empty')
if [ -z $hriApplicationId ]; then
  code=$(echo $response | jq -r '.code // empty')
  if [ $code == 'CONFLICT' ]; then
    echo
    echo 'App Id already configured! Continuing deployment without continuing App Id initialization.'

    # Get the existing applicationId to export as JWT_AUDIENCE_ID
    response=$(curl -X GET -sS "${issuer}/applications" -H "Authorization: Bearer ${iamToken}")
    hriApplicationId=$(echo $response | jq -r --arg name "$hriApplicationName" '.applications[] | select(.name == $name) | .clientId')

    if [ -z $hriApplicationId ]; then
      echo "Failed to get existing HRI Management API application ID! Unable to set JWT_AUDIENCE_ID!"
      echo "the_curl_response: $response"
      exit 1
    fi
    echo
    echo "Setting JWT_AUDIENCE_ID to existing HRI Management API ID: $hriApplicationId"
    echo $hriApplicationId > JWT_AUDIENCE_ID
    exit 0
  else
    echo "Failed to create ${APPID_PREFIX}HRI Management API application! Exit code ${code}."
    echo "the_curl_response: $response"
    exit 1
  fi
fi
echo $hriApplicationId > JWT_AUDIENCE_ID

# Assign scopes to application
echo
echo "Assigning hri_internal, hri_consumer and hri_data_integrator scopes to HRI provider application"
curl -X PUT -sS -f "${issuer}/applications/${hriApplicationId}/scopes" -H "Content-Type: application/json" -H "Authorization: Bearer ${iamToken}" -d @- << EOF
{
"scopes": [ "hri_internal", "hri_consumer", "hri_data_integrator"]
}
EOF

# Create roles
echo
echo "Creating roles for each of the created scopes"
response=$(curl -X POST -sS "${issuer}/roles" -H "Authorization: Bearer ${iamToken}" -H "Content-Type: application/json" -d @- << EOF
{
"name": "${APPID_PREFIX}HRI Internal",
"description": "HRI Internal Role",
"access": [ {
  "application_id": "${hriApplicationId}",
  "scopes": [ "hri_internal" ]
} ]
}
EOF
)
internalRoleId=$(echo $response | jq -r '.id // "REQUEST_FAILED"')
if [ $internalRoleId = "REQUEST_FAILED" ]; then
  echo "Error Creating role: HRI Internal!"
  echo "the_curl_response: $response"
  exit 1
fi

response=$(curl -X POST -sS "${issuer}/roles" -H "Authorization: Bearer ${iamToken}" -H "Content-Type: application/json" -d @- << EOF
{
"name": "${APPID_PREFIX}HRI Consumer",
"description": "HRI Consumer Role",
"access": [ {
  "application_id": "${hriApplicationId}",
  "scopes": [ "hri_consumer" ]
} ]
}
EOF
)
consumerRoleId=$(echo $response | jq -r '.id // "REQUEST_FAILED"')
if [ $consumerRoleId = "REQUEST_FAILED" ]; then
  echo "Error Creating role: HRI Consumer Role!"
  echo "the_curl_response: $response"
  exit 1  
fi

response=$(curl -X POST -sS "${issuer}/roles" -H "Authorization: Bearer ${iamToken}" -H "Content-Type: application/json" -d @- << EOF
{
"name": "${APPID_PREFIX}HRI Data Integrator",
"description": "HRI Data Integrator Role",
"access": [ {
  "application_id": "${hriApplicationId}",
  "scopes": [ "hri_data_integrator" ]
} ]
}
EOF
)
dataIntegratorRoleId=$(echo $response | jq -r '.id // "REQUEST_FAILED"')
if [ $dataIntegratorRoleId = "REQUEST_FAILED" ]; then
  echo "Error Creating role: HRI Data Integrator Role!"
  echo "the_curl_response: $response"
  exit 1  
fi

# Create HRI Internal application.
echo
echo "Creating HRI Internal application"
response=$(curl -X POST -sS "${issuer}/applications" -H "Authorization: Bearer ${iamToken}" -H 'Content-Type: application/json' -d @- << EOF
{
"name": "${APPID_PREFIX}HRI Internal",
"type": "regularwebapp"
}
EOF
)
internalApplicationId=$(echo $response | jq -r '.clientId // "REQUEST_FAILED"')
if [ $internalApplicationId = "REQUEST_FAILED" ]; then
  echo "Error Creating role: HRI Internal App Role!"
  echo "the_curl_response: $response"
  exit 1  
fi

# Assign roles to internal application.
echo
echo "Assigning internal and consumer roles to HRI Internal application"
curl -X PUT -sS -f "${issuer}/applications/${internalApplicationId}/roles" -H "Authorization: Bearer ${iamToken}" -H "Content-Type: application/json" -d @- << EOF
{
"roles":{
  "ids":["${internalRoleId}", "${consumerRoleId}"]
}}
EOF

exit 0
