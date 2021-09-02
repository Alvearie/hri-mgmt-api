#!/usr/bin/env bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

set -x

passing=0
failing=0
output=""

# lookup the base API url for the current targeted functions namespace
# Note: this doesn't work in MacOS due to differences in `sed` flags as compared to Linux.
serviceUrl=$(ibmcloud fn api list -f | grep 'URL: ' | grep 'hri/healthcheck' -m 1 | sed -rn 's/^.*: (.*)\/hri.*/\1\/hri/p')

echo 'Run Smoke Tests'

HRI_API_STATUS=$(curl --write-out "%{http_code}\n" --silent --output /dev/null "$serviceUrl/healthcheck" )
if [ $HRI_API_STATUS -eq 200 ]; then
  passing=$((passing+1))
  failure='/>'
else
  failing=$((failing+1))
  HRI_API_ERROR=$(curl "$serviceUrl/healthcheck")
  failure="><failure message=\"Expected HRI API healthcheck status to return code 200\" type=\"FAILURE\">$HRI_API_ERROR</failure></testcase>"
fi
output="$output\n<testcase classname=\"HRI-API\" name=\"GET Healthcheck\" time=\"0\"${failure}"

echo
echo "----------------"
echo "Final Results"
echo "----------------"
echo "PASSING: $passing"
echo "FAILING: $failing"
total=$(($passing + $failing))

date=`date`
header="<testsuite name=\"Smoke tests\" tests=\"$total\" failures=\"$failing\" errors=\"$failing\" skipped=\"0\" timestamp=\"${date}\" time=\"0\">"
footer="</testsuite>"

filename="smoketests.xml"
cat << EOF > $filename
$header
$output
$footer
EOF

if [ $failing -gt 0 ]; then
  exit 1
fi
