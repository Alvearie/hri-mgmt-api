#!/usr/bin/env bash

# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

set -x

passing=0
failing=0
output=""

./src/hri -config-path=test/spec/test_config/valid_config.yml >/dev/null &
sleep 1
HRI_WEB_SERVER_STATUS=$(curl -k --write-out "%{http_code}\n" --silent "$HRI_URL/healthcheck" )
if [ $HRI_WEB_SERVER_STATUS -eq 200 ]; then
  passing=$((passing+1))
  failure='/>'
else
  failing=$((failing+1))
  HRI_API_ERROR=$(curl "$HRI_URL/healthcheck")
  failure="><failure message=\"Expected HRI Web Server healthcheck status to return code 200\" type=\"FAILURE\">$HRI_API_ERROR</failure></testcase>"
fi
PROCESS_ID=$(lsof -iTCP:1323 -sTCP:LISTEN | grep -o '[0-9]\+' | sed 1q)
kill $PROCESS_ID

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
