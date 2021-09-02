# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

class AppIDHelper

  def initialize
    @helper = Helper.new
    @app_id_url = ENV['APPID_URL']
    @iam_token = IAMHelper.new.get_access_token
  end

  def get_access_token(application_name, scopes, audience_override = nil)
    response = @helper.rest_get("#{@app_id_url}/management/v4/#{ENV['APPID_TENANT']}/applications", {'Authorization' => "Bearer #{@iam_token}"})
    raise 'Failed to get AppId credentials' unless response.code == 200
    JSON.parse(response.body)['applications'].each do |application|
      if application['name'] == application_name
        @credentials = Base64.encode64("#{application['clientId']}:#{application['secret']}").delete("\n")
        break
      end
    end

    if @credentials.nil?
      raise "Unable to get AppID Application credentials for #{application_name}"
    end

    response = @helper.rest_post("#{@app_id_url}/oauth/v4/#{ENV['APPID_TENANT']}/token", {'grant_type' => 'client_credentials', 'scope' => scopes, 'audience' => (audience_override.nil? ? ENV['JWT_AUDIENCE_ID'] : audience_override)}, {'Content-Type' => 'application/x-www-form-urlencoded', 'Accept' => 'application/json', 'Authorization' => "Basic #{@credentials}"})
    raise 'App ID token request failed' unless response.code == 200
    JSON.parse(response.body)['access_token']
  end

end
