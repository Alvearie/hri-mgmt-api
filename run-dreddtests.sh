#!/bin/bash

npm install -g api-spec-converter
npm install -g dredd@12.2.0
gem install dredd_hooks

echo 'Clone Alvearie/hri-api-spec Repo'
git clone https://github.com/Alvearie/hri-api-spec.git api-spec
cd api-spec
echo "if exists, checkout ${TRAVIS_BRANCH}"
exists=$(git show-ref refs/remotes/origin/${TRAVIS_BRANCH})
if [[ -n "$exists" ]]; then
  git checkout ${TRAVIS_BRANCH}
else
  git checkout master
fi

# convert API to swagger 2.0
api-spec-converter -f openapi_3 -t swagger_2 -s yaml management-api/management.yml > management.swagger.yml
tac ../api-spec/management.swagger.yml | sed "1,8d" | tac > tmp && mv tmp ../api-spec/management.swagger.yml

# lookup the base API url for the current targeted functions namespace
serviceUrl=$(bx fn api list -f | grep 'URL: ' | grep -v batchId -m 1 | sed -rn 's/^.*: (.*)\/hri.*/\1\/hri/p')
dredd -r xunit -o ../dreddtests.xml management.swagger.yml $serviceUrl --sorted --language=ruby --hookfiles=../test/spec/dredd_hooks.rb --hooks-worker-connect-timeout=5000
