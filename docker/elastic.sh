#!/bin/bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

echo "Looking up ElasticSearch connection information and credentials"
# Get Elastic connection info
id=$(ibmcloud resource service-instance "${ELASTIC_INSTANCE}" --output json | jq .[0].id)
echo
echo "ES id: $id"

ibmcloud resource service-key "${ELASTIC_SVC_ACCOUNT}" --output json | jq ".[] | select(.source_crn == $id) | .credentials.connection.https.certificate.certificate_base64" | tr -d '"' | base64 -d > elastic.crt

export CURL_CA_BUNDLE=elastic.crt

baseUrl=$(ibmcloud resource service-key "${ELASTIC_SVC_ACCOUNT}" --output json | jq ".[] | select(.source_crn == $id) | .credentials.connection.https.composed[0]" | tr -d '"')
echo
# Remove the credentials from the url when logging
echo "ES baseUrl: ${baseUrl/:\/\/*@/://}"

# Keep track of any failures
rtn=0

# set auto-index creation off
echo "Setting ElasticSearch auto index creation to false"
curl -sS -f -X PUT $baseUrl/_cluster/settings  -H 'Content-Type: application/json' -d'
{
"persistent": { "action.auto_create_index": "false" }
}' || { echo -e 'Setting ElasticSearch auto index creation failed!' ; rtn=1; }

# upload batches index template
echo
echo -e "Setting ElasticSearch Batches index template"
curl -sS -f -X PUT $baseUrl/_index_template/batches -H 'Content-Type: application/json' -d '@batches.json' || 
{ 
	echo -e 'Setting ElasticSearch Batches index template failed!' ; rtn=1; 
}

echo
echo -e "ElasticSearch configuration complete"

exit $rtn
