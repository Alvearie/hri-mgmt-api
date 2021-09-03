#!/usr/bin/env ruby

# (C) Copyright IBM Corp. 2021
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

logger = Logger.new(STDOUT)

if %w[main develop].include?(ENV['TRAVIS_BRANCH'])
  logger.info("#{ARGV[0]} tests failed. Sending a message to Slack...")
  HRITestHelpers::SlackHelper.new(ENV['SLACK_WEBHOOK']).send_slack_message(ARGV[0], ENV['TRAVIS_BUILD_DIR'], ENV['TRAVIS_BRANCH'], ENV['TRAVIS_JOB_WEB_URL'])
else
  logger.info("#{ARGV[0]} tests failed, but a Slack message is only sent for the 'main' or 'develop' branches.")
end

exit 1