#!/usr/bin/env ruby
require_relative '../env'

logger = Logger.new(STDOUT)

if ENV['TRAVIS_BRANCH'] == 'master'
  logger.info("#{ARGV[0]} tests failed. Sending a message to Slack...")
  HRITestHelpers::SlackHelper.new(ENV['SLACK_WEBHOOK']).send_slack_message(ARGV[0], ENV['TRAVIS_BUILD_DIR'], ENV['TRAVIS_BRANCH'], ENV['TRAVIS_JOB_WEB_URL'])
else
  logger.info("#{ARGV[0]} tests failed, but a Slack message is only sent for the 'master' branch.")
end

exit 1