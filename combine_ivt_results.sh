#!/usr/bin/env bash

function combineTestResults() {
  INPUT_DIRECTORY="${1}"
  RESULT_FILE=$2
  SUITE_NAME="hri-mgmt-api - $BRANCH_NAME - IVT"
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
  header="<testsuite name=\"$SUITE_NAME\" tests=\"$testCount\" failures=\"$failures\" errors=\"$errors\" skipped=\"${skipped}\" timestamp=\"${date}\" time=\"${time}\">"
  footer="</testsuite>"
  echo -e "$header\n$output\n$footer" > "$RESULT_FILE"

  echo "Finished combining IVT JUnit files"
}

echo $(combineTestResults 'test/ivt_test_results' 'ivttest.xml')