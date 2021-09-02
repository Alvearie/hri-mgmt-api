# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

class SlackHelper

  def initialize
    @helper = Helper.new
    @slack_url = ENV['SLACK_WEBHOOK']
  end

  def send_slack_message(test_type)
    message = {
      text: "*#{test_type} Test Failure:*
              Repository: #{ENV['TRAVIS_BUILD_DIR'].split('/').last},
              Branch: #{ENV['TRAVIS_BRANCH']},
              Time: #{(Time.now - 14400).strftime("%m/%d/%Y %H:%M")},
              Build Link: #{ENV['TRAVIS_JOB_WEB_URL'].gsub('https:///', 'https://travis.ibm.com/')}"
    }.to_json
    @helper.rest_post(@slack_url, message)
  end

end