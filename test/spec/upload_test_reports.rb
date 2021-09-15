#!/usr/bin/env ruby

# (C) Copyright IBM Corp. 2021
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

# This script uploads JUnit test reports to Cloud Object Storage to be used by the Allure application to generate HTML
# test trend reports for the IVT and Dredd tests. More information on Allure can be found here: https://github.com/allure-framework/allure2
#
# The 'ivttest.xml' and 'dreddtests.xml' JUnit reports are uploaded to the 'wh-hri-dev1-allure-reports' Cloud Object Storage bucket,
# which is also mounted on the 'allure' kubernetes pod. This bucket keeps 30 days of reports that will be used to generate a
# historical HTML report when the allure executable is invoked on the pod.

cos_helper = HRITestHelpers::COSHelper.new(ENV['COS_URL'], ENV['IAM_CLOUD_URL'], ENV['CLOUD_API_KEY'])
logger = Logger.new(STDOUT)
time = Time.now.strftime '%Y%m%d%H%M%S'

if %w[main develop WHFHRI-667].include?(ENV['BRANCH_NAME'])
  if ARGV[0] == 'IVT'
    logger.info('Uploading ivttest.xml to COS')
    puts Dir.entries('.')
    puts '---------'
    puts Dir.entries('test/ivt_test_results')
    `sed -i 's#test/ivt_test_results#rspec#g' ivttest.xml`
    cos_helper.upload_object_data('wh-hri-dev1-allure-reports', "mgmt-api/#{ENV['BRANCH_NAME']}/ivt/ivttest.xml", File.read(File.join(Dir.pwd, "ivttest.xml")))
  elsif ARGV[0] == 'Dredd'
    logger.info('Uploading dreddtests.xml to COS')
    cos_helper.upload_object_data('wh-hri-dev1-allure-reports', "mgmt-api/#{ENV['BRANCH_NAME']}/dredd/dreddtests-#{time}.xml", File.read(File.join(Dir.pwd, 'dreddtests.xml')))
  else
    raise "Invalid argument: #{ARGV[0]}. Valid arguments: 'IVT' or 'Dredd'"
  end
else
  logger.info("Test reports are only generated for the 'main', 'develop', or 'WHFHRI-667' branches. Exiting.")
end