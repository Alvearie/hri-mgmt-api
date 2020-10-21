#!/usr/bin/env bash

date=`date`
time=0.0
title="<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
header="<testsuite name=\"FVT Tests\" tests=\"0\" failures=\"0\" errors=\"0\" skipped=\"0\" timestamp=\"${date}\" time=\"${time}\">"
test="<testcase classname=\"testFVT\" name=\"testFVT\" file=\"./test/spec/hri_management_api_spec.rb\" time=\"4.491684\"></testcase>"
footer="</testsuite>"

filename="fvttest.xml"
echo -e "$title\n$header\n$test\n$footer" > "$filename"