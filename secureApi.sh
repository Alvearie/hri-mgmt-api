#!/bin/bash

# This script enable API key authentication to an API. It downloads the API swagger definition,
# adds a security definition, and then updates the API with the new configuration.  See these
# for more detailed information:
# https://github.com/apache/openwhisk/blob/master/docs/apigateway.md
# https://github.com/apache/openwhisk-apigateway/blob/master/doc/v2/management_interface_v2.md

# Download the api definition
ibmcloud fn api get hri-batches > api.json

# check for an existing security definition
if [[ -n $(grep securityDefinitions api.json) ]]; then
  echo "Found existing security definitions, exiting"
  exit 0
fi

# Remove the last two lines
lines=$(wc -l < api.json)
head -n $(( lines - 2 )) api.json > secureApi.json

# Append the security definition and closing '}'
cat >>secureApi.json <<EOL
    },
    "securityDefinitions": {
        "client_id": {
            "in": "header",
            "name": "X-IBM-Client-Id",
            "type": "apiKey",
            "x-key-type": "clientId"
        }
    },
    "security": [
        {
            "client_id": []
        }
    ]
}
EOL

# upload the new config
ibmcloud fn api create --config-file secureApi.json
