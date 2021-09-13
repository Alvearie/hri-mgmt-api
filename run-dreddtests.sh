#!/bin/bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

sudo npm install -g api-spec-converter
sudo npm install -g dredd@12.2.0

echo 'Clone Alvearie/hri-api-spec Repo'
git clone https://github.com/Alvearie/hri-api-spec.git hri-api-spec
cd hri-api-spec
echo "if exists, checkout ${BRANCH_NAME}"
exists=$(git show-ref refs/remotes/origin/${BRANCH_NAME})
if [[ -n "$exists" ]]; then
  git checkout ${BRANCH_NAME}
elif [ -n "$API_SPEC_TAG" ]; then
  git checkout -b mgmt-api_auto_dredd $API_SPEC_TAG
else
  git checkout $API_SPEC_DEV_BRANCH
fi

#Convert API to swagger 2.0
api-spec-converter -f openapi_3 -t swagger_2 -s yaml management-api/management.yml > management.swagger.yml
tac ../hri-api-spec/management.swagger.yml | sed "1,8d" | tac > tmp && mv tmp ../hri-api-spec/management.swagger.yml

# lookup the base API url for the current targeted functions namespace
serviceUrl=$(bx fn api list -f | grep 'URL: ' | grep -v batchId -m 1 | sed -rn 's/^.*: (.*)\/hri.*/\1\/hri/p')
dredd -r xunit -o ../dreddtests.xml management.swagger.yml $serviceUrl --sorted --language=ruby --hookfiles=../test/spec/dredd_hooks.rb --hooks-worker-connect-timeout=5000
