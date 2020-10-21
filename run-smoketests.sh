#!/usr/bin/env bash
set -x
# Requires an HRI_API_KEY if authorization is enabled

passing=0
failing=0
output=""

# Acquire Service Endpoint
healthcheckAction=$(ibmcloud fn action list | awk '/healthcheck/{print $1}')
apiHost=$(ibmcloud fn property get --apihost | awk '{print $4}')
healthcheckUrl="https://$apiHost/api/v1/web$healthcheckAction"

echo 'Run Smoke Tests'

# Don't display FN_WEB_SECURE_KEY
set +x
HRI_API_STATUS=$(curl --write-out "%{http_code}\n" --silent --output /dev/null "$healthcheckUrl" -H "X-Require-Whisk-Auth: $FN_WEB_SECURE_KEY")
set -x
if [ $HRI_API_STATUS -eq 200 ]; then
  passing=$((passing+1))
  failure='/>'
else
  failing=$((failing+1))
  HRI_API_ERROR=$(curl "$healthcheckUrl" -H "X-Require-Whisk-Auth: ${FN_WEB_SECURE_KEY}")
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
