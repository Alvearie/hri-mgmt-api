#!/usr/bin/env ruby
# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

# This script uploads JUnit test reports to Cloud Object Storage to be used by the UnitTH application to generate HTML
# test trend reports for the IVT and Dredd tests. More information on unitth can be found here: http://junitth.sourceforge.net/
#
# The 'ivttest.xml' and 'dreddtests.xml' JUnit reports are uploaded to the 'hri-test-reports' Cloud Object Storage bucket,
# which is also mounted on the 'unitth' kubernetes pod. This bucket keeps 30 days of reports that will be used to generate a
# historical HTML report when the UnitTH jar is run on the pod.

cos_helper = COSHelper.new
logger = Logger.new(STDOUT)
time = Time.now.strftime '%Y%m%d%H%M%S'

if ENV['TRAVIS_BRANCH'] == ARGV[0]
  if ARGV[1] == 'IVT'
    logger.info('Uploading ivttest.xml to COS')
    `sed -i 's#test/ivt_test_results#rspec#g' ivttest.xml`
    cos_helper.upload_object_data('wh-hri-dev1-test-reports', "mgmt-api/#{ENV['TRAVIS_BRANCH']}/ivt/#{time}/ivttest.xml", File.read(File.join(Dir.pwd, 'ivttest.xml')))
  elsif ARGV[1] == 'Dredd'
    logger.info('Uploading dreddtests.xml to COS')
    cos_helper.upload_object_data('wh-hri-dev1-test-reports', "mgmt-api/#{ENV['TRAVIS_BRANCH']}/dredd/#{time}/dreddtests.xml", File.read(File.join(Dir.pwd, 'dreddtests.xml')))
  else
    raise "Invalid argument: #{ARGV[1]}. Valid arguments: 'IVT' or 'Dredd'"
  end
else
  logger.info("Test reports are only generated for the #{ARGV[0]} branch. Exiting.")
end