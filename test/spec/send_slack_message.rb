#!/usr/bin/env ruby
# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

logger = Logger.new(STDOUT)

if ENV['TRAVIS_BRANCH'] == 'release-2.1-2.1'
  logger.info("#{ARGV[0]} tests failed. Sending a message to Slack...")
  SlackHelper.new.send_slack_message(ARGV[0])
else
  logger.info("#{ARGV[0]} tests failed, but a Slack message is only sent for the release-2.1-2.1 branch.")
end

exit 1