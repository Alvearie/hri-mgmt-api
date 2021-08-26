#!/usr/bin/env bash

PUBLISH_TYPE=$1

export MY_APP_NAME="hri-mgmt-api"

function combineTestResults() {
  echo "===============BEGIN COMBINING TESTS==============="

  INPUT_DIRECTORY="${1}"
  RESULT_FILE=$2
  failures=0
  testCount=0
  errors=0
  skipped=0
  time=0.0
  output=""

  for file_name in ${INPUT_DIRECTORY}/*.xml; do
    newOutput=""
    newTime=0.0
    testCount=$((testCount+$(cat "$file_name" | grep -o 'tests="[^"]*' | sed 's/tests="//g')))
    failures=$((failures+$(cat "$file_name" | grep -o 'failures="[^"]*' | sed 's/failures="//g')))
    errors=$((errors+$(cat "$file_name" | grep -o 'errors="[^"]*' | sed 's/errors="//g')))
    skipped=$((skipped+$(cat "$file_name" | grep -o 'skipped="[^"]*' | sed 's/skipped="//g')))
    newTime=$(cat "$file_name" | head -2 | tail -1 | grep -o 'time="[^"]*' | sed 's/time="//g')
    time=$(awk "BEGIN {print $time+$newTime; exit}")
    newOutput=$(cat "$file_name" | tail -n +3 | sed '$d')
    output="$output$newOutput"
  done

  date=`date`
  header="<testsuite name=\"$INPUT_DIRECTORY\" tests=\"$testCount\" failures=\"$failures\" errors=\"$errors\" skipped=\"${skipped}\" timestamp=\"${date}\" time=\"${time}\">"
  footer="</testsuite>"
  echo -e "$header\n$output\n$footer" > "$RESULT_FILE"

  echo "===============END COMBINING TESTS==============="
}

function ifFileExists() {
  FILE_PATH=$1
  if [ -f "$FILE_PATH" ]
  then
    echo "Found test file to publish. Continuing..."
  else
    echo "Cannot find file path $FILE_PATH"
  exit 1
  fi
}

if [ "$PUBLISH_TYPE" == "buildRecord" ]
then
  # Upload a build record for this build, It is assumed that the build was successful
  ibmcloud doi publishbuildrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --branch $TRAVIS_BRANCH --repositoryurl https://github.ibm.com/wffh-hri/mgmt-api --commitid $TRAVIS_COMMIT --status pass

elif [ "$PUBLISH_TYPE" == "deployRecord" ]
then
  # Upload a deployment record; It is assumed that the deployment was successful
  ibmcloud doi publishdeployrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --env=dev --status=pass

elif [ "$PUBLISH_TYPE" == "unitTest" ]
then
  # Upload unittest test record for the build
  ifFileExists "unittest.xml"
  ibmcloud doi publishtestrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --filelocation=unittest.xml --type=unittest

elif [ "$PUBLISH_TYPE" == "ivtTest" ]
then
  # Upload IVT test record for the build
  echo $(combineTestResults 'test/ivt_test_results' 'ivttest.xml')
  ifFileExists "ivttest.xml"
  ibmcloud doi publishtestrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --filelocation=ivttest.xml --type=ivt

elif [ "$PUBLISH_TYPE" == "dreddTest" ]
then
  # Upload Dredd test record as IVT for the build
  ifFileExists "dreddtests.xml"
  ibmcloud doi publishtestrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --filelocation=dreddtests.xml --type=ivt

elif [ "$PUBLISH_TYPE" == "fvtTest" ]
then
  # Upload FVT test record for the build
  ifFileExists "fvttest.xml"
  ibmcloud doi publishtestrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --filelocation=fvttest.xml --type=fvt

elif [ "$PUBLISH_TYPE" == "smokeTest" ]
then
  # Upload Smoke test record for the build
  echo $(combineTestResults 'build/test-results/smokeTest' 'smoketests.xml')
  ifFileExists "smoketests.xml"
  ibmcloud doi publishtestrecord --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --filelocation=smoketests.xml --type=smoketests

elif [ "$PUBLISH_TYPE" == "sonarQube" ]
then
  # Upload SonarQube test record for the build
  REPORT_TASK_PATH="build/sonar/report-task.txt"
  ifFileExists "$REPORT_TASK_PATH"
  ibmcloud doi publishtestrecord --sqtoken="$SONARQUBE_CREDENTIALS" --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --filelocation=$REPORT_TASK_PATH --type=sonarqube
  find . -name '*.json'
  ifFileExists "SQData_$MY_APP_NAME.json"

elif [ "$PUBLISH_TYPE" == "evaluateCI" ]
then
  # Invoke a DevOps Insights gate to evaluated a policy based on uploaded data
  ibmcloud doi evaluategate --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --policy=WFFH-CI

elif [ "$PUBLISH_TYPE" == "evaluateCD" ]
then
  # Invoke a DevOps Insights gate to evaluated a policy based on uploaded data
  ibmcloud doi evaluategate --logicalappname="$MY_APP_NAME" --buildnumber="$TRAVIS_BUILD_NUMBER" --policy=WFFH-CD
fi