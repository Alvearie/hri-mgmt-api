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

cos_helper = COSHelper.new
logger = Logger.new(STDOUT)
time = Time.now.strftime '%Y%m%d%H%M%S'

if ARGV[0] == 'IVT'
  logger.info("Uploading ivttest-#{time}.xml to COS")
  text = File.read("#{Dir.pwd}/ivttest.xml")
  text = text.gsub!('testsuite name="Dredd Tests"', %Q(testsuite name="hri-mgmt-api - #{ENV['BRANCH_NAME']} - Dredd"))
  File.rename("#{Dir.pwd}/ivttest.xml", "#{Dir.pwd}/hri-mgmt-api-ivttest-#{time}.xml")
  cos_helper.upload_object_data('wh-hri-dev1-allure-reports', "hri-mgmt-api-ivttest-#{time}.xml", File.read(File.join(Dir.pwd, "hri-mgmt-api-ivttest-#{time}.xml")))
elsif ARGV[0] == 'Dredd'
  logger.info("Uploading dreddtests-#{time}.xml to COS")
  File.open("#{Dir.pwd}/dreddtests.xml", "w") { |file| file.puts text }
  doc = Nokogiri::XML(File.open("#{Dir.pwd}/dreddtests.xml")) { |file| file.noblanks }
  doc.search('//testsuite').attribute('name').value = "hri-mgmt-api - #{ENV['BRANCH_NAME']} - Dredd"

  #Dredd XUnit output contains skipped tests, so delete these elements from the result xml
  doc.search('//testsuite/testcase').each do |block|
    block.remove if block.children.count == 1
  end

  File.write("#{Dir.pwd}/dreddtests.xml", doc)
  File.rename("#{Dir.pwd}/dreddtests.xml", "#{Dir.pwd}/hri-mgmt-api-dreddtests-#{time}.xml")
  cos_helper.upload_object_data('wh-hri-dev1-allure-reports', "hri-mgmt-api-dreddtests-#{time}.xml", File.read(File.join(Dir.pwd, "hri-mgmt-api-dreddtests-#{time}.xml")))
else
  raise "Invalid argument: #{ARGV[0]}. Valid arguments: 'IVT' or 'Dredd'"
end