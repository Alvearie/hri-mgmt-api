# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

describe 'HRI Management API Without Validation' do

  INVALID_ID = 'INVALID'
  TENANT_ID = 'test0211'
  AUTHORIZED_TENANT_ID = 'provider1234'
  TENANT_ID_WITH_NO_ROLES = 'aztest'
  TENANT_WITH_DATA_INTEGRATOR_ROLE = 'qatest'
  TENANT_WITH_DATA_CONSUMER_ROLE = 'provider237'
  TENANT_WITH_INTEGRATOR_CONSUMER_ROLE = 'test123'
  INTEGRATOR_ID = 'claims'
  TEST_TENANT_ID = "rspec-#{'-'.delete('.')}-test-tenant".downcase
  TEST_INTEGRATOR_ID = "rspec-#{'-'.delete('.')}-test-integrator".downcase
  DATA_TYPE = 'rspec-batch'
  INVALIDTHRESHOLD = 10
  STATUS = 'started'
  BATCH_INPUT_TOPIC = "ingest.#{AUTHORIZED_TENANT_ID}.#{INTEGRATOR_ID}.in"
  KAFKA_TIMEOUT = 60

  def initialize(mongodb_credentials = {})
    @headers = { 'Content-Type': 'application/json' }
    client = Mongo::Client.new('mongodb://hi-dp-tst-eastus-cosmos-mongo-api-hri:Jl6rN2wUFpROlr4Cxse61ET51TB1qwZTZXfD1IwotXQKUBUaEGjBXr8DqKAKonhBkhwSxdLIkJitZUE9X2liSg==@hi-dp-tst-eastus-cosmos-mongo-api-hri.mongo.cosmos.azure.com:10255/?ssl=true&replicaSet=globaldb&retrywrites=false&maxIdleTimeMS=120000&appName=@hi-dp-tst-eastus-cosmos-mongo-api-hri@' , :database => 'HRI-DEV')
    db = client.database
    collection = client[:'HRI-Mgmt']
    #puts collection.find( { tenantid: 'q-batches' } ).first
  end

  def get_client_id_and_secret()
    return ['c33ac4da-21c6-426b-abcc-27e24ff1ccf9','GxF8Q~XfZyLRQBZ4mjwgEogVWwGjtzJh7ZPzgagw']
  end

  def generate_access_token()
    credentials = get_client_id_and_secret
    response = rest_post("https://login.microsoftonline.com/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/oauth2/v2.0/token",{'grant_type' => 'client_credentials','scope' => 'c33ac4da-21c6-426b-abcc-27e24ff1ccf9/.default', 'client_secret' => 'GxF8Q~XfZyLRQBZ4mjwgEogVWwGjtzJh7ZPzgagw', 'client_id' => 'c33ac4da-21c6-426b-abcc-27e24ff1ccf9'}, {'Content-Type' => 'application/x-www-form-urlencoded', 'Accept' => 'application/json', 'Authorization' => "Basic #{Base64.encode64("#{credentials[0]}:#{credentials[1]}").delete("\n")}" })
    raise 'App ID token request failed' unless response.code == 200
    #puts "This is the generated token ==============================>  "
    #puts JSON.parse(response.body)['access_token']
    JSON.parse(response.body)['access_token']
  end

  def es_get_batch(index, batch_id)
    rest_get("#{@hri_base_url}/#{index}-batches/_doc/#{batch_id}", @headers, @basic_auth)
  end

  def es_batch_update(index, batch_id, script)
    # wait_for waits for an index refresh to happen, so increase the standard timeout
    rest_post("#{@hri_base_url}/#{index}-batches/_doc/#{batch_id}/_update?refresh=wait_for", script, @headers, @basic_auth)
  end

  def hri_post_tenant(tenant_id, request_body = nil, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_post(url, request_body, headers)
  end

  def hri_post_tenant_stream(tenant_id, integrator_id, request_body, override_headers = {}, delete_auth = false)
    url = "#{@hri_base_url}/tenants/#{tenant_id}/streams/#{integrator_id}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    headers.delete('Authorization') if delete_auth
    rest_post(url, request_body, headers, {})
  end

  def hri_delete_tenant(tenant_id, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_delete(url, nil, headers, {})
  end

  def hri_get_tenant_streams(tenant_id, override_headers = {}, delete_auth = false)
    url = "#{@hri_base_url}/tenants/#{tenant_id}/streams"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    headers.delete('Authorization') if delete_auth
    rest_get(url, headers)
  end


  def hri_get_tenant(tenant_id, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_get_tenants(override_headers = {})
    url = "#{@hri_base_url}/tenants"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_delete_tenant_stream(tenant_id, integrator_id, override_headers = {}, delete_auth = false)
    url = "#{@hri_base_url}/tenants/#{tenant_id}/streams/#{integrator_id}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    headers.delete('Authorization') if delete_auth
    rest_delete(url, nil, headers, {})
  end

  def hri_get_batch(tenant_id, batch_id, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches/#{batch_id}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_get_batches(tenant_id, query_params = nil, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches"
    url += "?#{query_params}" unless query_params.nil?
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_put_batch(tenant_id, batch_id, action, additional_params = {}, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches/#{batch_id}/action/#{action}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}"}.merge(override_headers)
    @request_helper.rest_put(url, additional_params.to_json, headers)
  end

  def hri_post_batch(tenant_id, request_body, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}"}.merge(override_headers)
    rest_post(url, request_body, headers)
  end

  def response_rescue_wrapper
    yield
  rescue Exception => e
    raise e unless defined?(e.response)
    logger_message(@hri_api_info, e)
    e.response
  end

  def rest_client_resource_for
    @hri_api_info[:verify_ssl] = OpenSSL::SSL::VERIFY_NONE

    response_rescue_wrapper do
      RestClient::Request.execute(@hri_api_info)
    end
  end

  def rest_get(url, headers, overrides = {})
    @hri_api_info = { method: :get, url: url, headers: headers }.merge(overrides)
    response_rescue_wrapper do
      rest_client_resource_for
    end
  end

  def rest_post(url, body, override_headers = {}, overrides = {})
    headers = { 'Accept' => '*/*' }.merge(override_headers)
    @hri_api_info = { method: :post, url: url, headers: headers, payload: body }.merge(overrides)
    response_rescue_wrapper do
      rest_client_resource_for
    end
  end

  def rest_put(url, body, override_headers = {}, overrides = {})
    headers = { 'Accept' => '*/*' }.merge(override_headers)
    @hri_api_info = { method: :put, url: url, headers: headers, payload: body }.merge(overrides)
    response_rescue_wrapper do
      rest_client_resource_for
    end
  end

  def rest_delete(url, body, override_headers = {}, overrides = {})
    headers = { 'Accept' => '*/*' }.merge(override_headers)
    @hri_api_info = { method: :delete, url: url, headers: headers, payload: body }.merge(overrides)
    response_rescue_wrapper do
      rest_client_resource_for
    end
  end

  def logger_message(info, error)
    printed_info = if info[:headers].nil?
                     info
                   else
                     headers = info[:headers].dup
                     headers['Authorization'] = headers['Authorization'].split(' ')[0] + ' [REDACTED]' if headers['Authorization']
                     info.merge(headers: headers)
                   end
    Logger.new(STDOUT).info("Received exception hitting endpoint: #{printed_info}. Exception: #{error}, response: #{error.response}")
  end

  def hri_custom_request(request_url, request_body = nil, override_headers = {}, request_type)
    url = "#{@hri_base_url}#{request_url}"
    @az_token = generate_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json' }.merge(override_headers)
    case request_type
    when 'GET'
      rest_get(url, headers)
    when 'PUT'
      rest_put(url, request_body, headers)
    when 'POST'
      headers['Authorization'] = "Bearer #{@az_token}"
      rest_post(url, request_body, headers)
    when 'DELETE'
      headers['Authorization'] = "Bearer #{@az_token}"
      rest_delete(url, nil, headers)
    else
      raise "Invalid request type: #{request_type}"
    end
  end


  def get_topics
    response = rest_get("#{@admin_url}/admin/topics", @headers)
    puts "================ Topics Created are =======================> "
    puts response
    raise 'Failed to get Event Streams topics' unless response.code == 200
    puts JSON.parse(response.body).map { |topic| topic['name']}
  end



  before(:all) do
    @hri_base_url = "https://hri-1.wh-wcm.dev.watson-health.ibm.com/hri"
    @request_helper = HRITestHelpers::RequestHelper.new
    @azure_token = generate_access_token()
    @mgmt_api_helper = HRITestHelpers::MgmtAPIHelper.new(@hri_base_url, @azure_token)
    @hri_deploy_helper = HRIDeployHelper.new
    @app_id_helper = HRITestHelpers::AppIDHelper.new(ENV['APPID_URL'], ENV['APPID_TENANT'], @iam_token, ENV['JWT_AUDIENCE_ID'])
    @start_date = DateTime.now


    @exe_path = File.absolute_path(File.join(File.dirname(__FILE__), "../../src/hri"))
    @config_path = File.absolute_path(File.join(File.dirname(__FILE__), "test_config"))
    @log_path = File.absolute_path(File.join(File.dirname(__FILE__), "../logs"))
    Dir.mkdir(@log_path) if !Dir.exists?(@log_path)


    #@hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path, 'no-validation-1-')
    response = rest_get("#{@hri_base_url}/healthcheck", {})
    unless response.code == 200
    raise "Health check failed: #{response.body}"
    end


    #Create Batch
    @batch_prefix = "batch"
    @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
    create_batch = {
      name: @batch_name,
      topic: BATCH_INPUT_TOPIC,
      status: STATUS,
      dataType: "#{DATA_TYPE}",
      invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
      metadata: {
        "compression": "gzip",
        "finalRecordCount": 20
      }
    }
    response = hri_post_batch(AUTHORIZED_TENANT_ID, create_batch.to_json)
    expect(response.code).to eq 201
    parsed_response = JSON.parse(response.body)
    @batch_id = parsed_response['id']
    puts parsed_response
    Logger.new(STDOUT).info("New Batch Created With ID: #{@batch_id}")

  end


  context 'POST /tenants/{tenant_id}' do

    it 'Success' do
      response = hri_post_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql TEST_TENANT_ID
      delete_tenant = hri_delete_tenant(TEST_TENANT_ID)
    end

    it 'Tenant Already Exists' do
      response = hri_post_tenant(TENANT_ID)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "Unable to create new tenant as it already exists[#{TENANT_ID}]: [400]"
    end

    it 'Invalid Tenant ID' do
      response = hri_post_tenant(INVALID_ID)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"
    end

    it 'Tenant ID containing special character' do
      response = hri_post_tenant("test@tenant")
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "invalid request arguments:\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"
    end

    it 'Tenant ID starting with underscore' do
      response = hri_post_tenant("_#{TENANT_ID}")
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "Unable to create a new tenant[_#{TENANT_ID}]:[400]"
    end

    it 'Tenant ID containing/ending with hyphon' do
      response = hri_post_tenant("#{TENANT_ID}-123")
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql "#{TENANT_ID}-123"
      delete_tenant = hri_delete_tenant("#{TENANT_ID}-123")
    end

    it 'Unauthorized' do
      response = hri_post_tenant(TEST_TENANT_ID, nil, { 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end


    it 'Missing Tenant ID' do
      response = hri_post_tenant(nil)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Tenant ID With Ending Forward Slash' do
      response = hri_custom_request('/tenants//', nil, {}, 'POST')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Tenant ID With No Forward Slash' do
      response = hri_custom_request('/tenants', nil, {}, 'POST')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'POST /tenants/{tenant_id}/streams/{integrator_id}' do

    before(:each) do
      @stream_info = {
        numPartitions: 1,
        retentionMs: 3600000,
        cleanupPolicy: 'delete'
      }
    end

    it 'Success' do
      #Create Stream
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 201

      #Verify Stream Creation
      response = hri_get_tenant_streams(TEST_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'][0]['id']).to eql TEST_INTEGRATOR_ID
      delete_stream = hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID)

    end

    it 'Stream Already Exists' do
      response = hri_post_tenant_stream(TENANT_ID, INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Topic 'ingest.#{TENANT_ID}.#{INTEGRATOR_ID}.in' already exists."
    end

    it 'Missing numPartitions' do
      @stream_info.delete(:numPartitions)
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- numPartitions (json field in request body) is a required field"
    end

    it 'Invalid numPartitions' do
      @stream_info[:numPartitions] = '1'
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"numPartitions\": expected type int64, but received type string"
    end

    it 'Missing retentionMs' do
      @stream_info.delete(:retentionMs)
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- retentionMs (json field in request body) is a required field"
    end

    it 'Invalid retentionMs' do
      @stream_info[:retentionMs] = '3600000'
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"retentionMs\": expected type int, but received type string"
    end

    it 'Invalid Stream Name' do
      response = hri_post_tenant_stream(TEST_TENANT_ID, ".#{TEST_INTEGRATOR_ID}.#{TEST_INTEGRATOR_ID}", @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'"
    end

    it 'Invalid Stream Name containing two dots' do
      response = hri_post_tenant_stream(TEST_TENANT_ID, ".#{TEST_INTEGRATOR_ID}.#{TEST_INTEGRATOR_ID}.#{TEST_INTEGRATOR_ID}", @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'"
    end

    it 'Invalid cleanupPolicy' do
    @stream_info[:cleanupPolicy] = 12345
    response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
    expect(response.code).to eq 400
    parsed_response = JSON.parse(response.body)
    expect(parsed_response['errorDescription']).to eql "invalid request param \"cleanupPolicy\": expected type string, but received type number"
    end

    it 'cleanupPolicy must be "compact" or "delete"' do
    @stream_info[:cleanupPolicy] = "invalid"
    response = hri_post_tenant_stream(TEST_TENANT_ID, 'test', @stream_info.to_json)
    expect(response.code).to eq 400
    parsed_response = JSON.parse(response.body)
    expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- cleanupPolicy (json field in request body) must be one of [delete compact]"
    end

    it 'Invalid cleanupPolicy and missing numPartitions' do
      @stream_info[:cleanupPolicy] = INVALID_ID
      @stream_info.delete(:numPartitions)
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- cleanupPolicy (json field in request body) must be one of [delete compact]\n- numPartitions (json field in request body) is a required field"
    end

    it 'Missing Authorization' do
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json, {'Authorization' => nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Azure AD authentication returned 401"
    end

    it 'Invalid Authorization' do
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json, {'Authorization' => 'Bearer Invalid'})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Missing Tenant ID' do
      response = hri_post_tenant_stream(nil, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Stream ID' do
      response = hri_post_tenant_stream(TEST_TENANT_ID, nil, @stream_info.to_json)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Stream ID With Ending Forward Slash' do
      response = hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams//", @stream_info.to_json, {}, 'POST')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Stream ID With No Forward Slash' do
      response = hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams", @stream_info.to_json, {}, 'POST')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'DELETE /tenants/{tenant_id}/streams/{integrator_id}' do

    before(:each) do
          @stream_info = {
            numPartitions: 1,
            retentionMs: 3600000,
            cleanupPolicy: 'delete'
          }
    end

    it 'Success' do
      #Delete Stream and Verify Deletion
      loop do
        create_stream = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
        response = hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID)
        break if response.code == 200

        response = hri_get_tenant_streams(TEST_TENANT_ID)
        expect(response.code).to eq 200
        parsed_response = JSON.parse(response.body)
        break if parsed_response['results'] == []
        sleep 1
      end

    end

    it 'Invalid Stream' do
      response = hri_delete_tenant_stream(INVALID_ID, TEST_INTEGRATOR_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unable to delete topic \"ingest.INVALID.rspec---test-integrator.in\": Broker: Unknown topic or partition\nUnable to delete topic \"ingest.INVALID.rspec---test-integrator.notification\": Broker: Unknown topic or partition"
    end

    it 'Invalid Stream containing two dots' do
      response = hri_delete_tenant_stream(INVALID_ID, "#{TEST_INTEGRATOR_ID}.#{TEST_INTEGRATOR_ID}.#{TEST_INTEGRATOR_ID}")
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'"
    end

    it 'Missing Tenant ID' do
      response = hri_delete_tenant_stream(nil, INTEGRATOR_ID)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Authorization' do
      response = hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, {'Authorization' => nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Azure AD authentication returned 401"
    end

    it 'Invalid Authorization' do
      response = hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, {'Authorization' => 'Bearer Invalid'})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Azure AD authentication returned 401"
    end

    it 'Missing Stream ID' do
      response = hri_delete_tenant_stream(TEST_TENANT_ID, nil)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Stream ID With Ending Forward Slash' do
      response = hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams//", nil, {}, 'DELETE')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Stream ID With No Forward Slash' do
      response = hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams", nil, {}, 'DELETE')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'DELETE /tenants/{tenant_id}' do

    it 'Success' do
      #Delete Tenant
      create_tenant = hri_post_tenant(TEST_TENANT_ID)
      response = hri_delete_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 200

      #Verify Tenant Deleted
      response = hri_get_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql"Tenant: #{TEST_TENANT_ID} not found: [404]"

    end

    it 'Invalid Tenant ID' do
      response = hri_delete_tenant(INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Could not delete tenant [#{INVALID_ID}-batches]: [404]"
    end

    it 'Unauthorized - No bearer token passed' do
      response = hri_delete_tenant(TEST_TENANT_ID, { 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Bearer token passed' do
      response = hri_delete_tenant(TEST_TENANT_ID, { 'Authorization': "Invalid"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Missing Tenant ID' do
      response = hri_delete_tenant(nil)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Tenant ID With Ending Forward Slash' do
      response = hri_custom_request('/tenants//', nil, {}, 'DELETE')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Tenant ID With No Forward Slash' do
      response = hri_custom_request('/tenants', nil, {}, 'DELETE')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'GET /tenants/{tenant_id}/streams' do

    it 'Success' do
      response = hri_get_tenant_streams(TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'][0]['id']).to eql INTEGRATOR_ID
    end

    it 'Success With Invalid Topic Only' do
      response = hri_get_tenant_streams(TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      stream_found = false
      parsed_response['results'].each do |integrator|
        stream_found = true if integrator['id'] == INTEGRATOR_ID
      end
      raise "Tenant Stream Not Found: #{INTEGRATOR_ID}" unless stream_found
    end

    it 'Missing Tenant ID' do
      response = hri_get_tenant_streams(nil)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Authorization' do
      response = hri_get_tenant_streams(TENANT_ID, {'Authorization' => nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Azure AD authentication returned 401"
    end

    it 'Invalid Authorization' do
      response = hri_get_tenant_streams(TENANT_ID, {'Authorization' => 'Bearer Invalid'})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

  end


  context 'GET /tenants' do

    it 'Success' do
      response = hri_get_tenants
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].to_s).to include TENANT_ID
    end

    it 'Unauthorized - No bearer token passed' do
      response = hri_get_tenants({ 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid bearer token passed' do
      response = hri_get_tenants({ 'Authorization': "Invalid token" })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

  end

  context 'GET /tenants/{tenant_id}' do

    it 'Success' do
      create_tenant = hri_post_tenant(TEST_TENANT_ID)
      response = hri_get_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql "#{TEST_TENANT_ID}-batches"
      expect(parsed_response['health']).to eql 'green'
      expect(parsed_response['status']).to eql 'open'
      delete_tenant = hri_delete_tenant(TEST_TENANT_ID)
    end

    it 'Invalid Tenant ID' do
      response = hri_get_tenant(INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Tenant: #{INVALID_ID} not found: [404]"
    end

    it 'Unauthorized- No bearer token passed' do
      response = hri_get_tenant(INVALID_ID, {'Authorization': nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized- Invalid bearer token passed' do
      response = hri_get_tenant(INVALID_ID, {'Authorization': "Invalid token"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Ending Forward Slash' do
      response = hri_custom_request('/tenants//', nil, {}, 'GET')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

  end

  context 'POST /tenants/{tenant_id}/batches' do

    before(:each) do
      @batch_prefix = 'batch'
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        topic: BATCH_INPUT_TOPIC,
        dataType: "#{DATA_TYPE}",
        invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }
    end

    it 'Successful Batch Creation' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Verify Batch in DB
      response_val = hri_get_batch(AUTHORIZED_TENANT_ID, @new_batch_id)
      parsed_response = JSON.parse(response_val.body)
      expect(parsed_response['id']).to eql @new_batch_id
      expect(parsed_response['name']).to eql @batch_name
      expect(parsed_response['status']).to eql "started"
      expect(parsed_response['topic']).to eql BATCH_INPUT_TOPIC
      expect(parsed_response['dataType']).to eql "#{DATA_TYPE}"
      expect(parsed_response['invalidThreshold']).to eql "#{INVALIDTHRESHOLD}".to_i
      expect(parsed_response['metadata']['compression']).to eql 'gzip'
      expect(parsed_response['metadata']['finalRecordCount']).to eql 20
      expect(DateTime.parse(parsed_response['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")

    end

    it 'Batch creation with an invalid Kafka topic' do
      #Create Batch with Bad Topic
      @batch_template[:topic] = 'INVALID-TEST-TOPIC'
      @batch_template[:dataType] = 'rspec-invalid-batch'
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 504

    end

    it 'Invalid Name' do
      @batch_template[:name] = 12345
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"name\": expected type string, but received type number"
    end

    it 'Invalid Topic' do
      @batch_template[:topic] = 12345
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"topic\": expected type string, but received type number"
    end

    it 'Invalid Data Type' do
      @batch_template[:dataType] = 12345
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"dataType\": expected type string, but received type number"
    end

    it 'Missing Name' do
      @batch_template.delete(:name)
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- name (json field in request body) is a required field"
    end

    it 'Missing Topic' do
      @batch_template.delete(:topic)
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- topic (json field in request body) is a required field"
    end

    it 'Missing Data Type' do
      @batch_template.delete(:dataType)
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- dataType (json field in request body) is a required field"
    end

    it 'Missing Name, Topic, and Data Type' do
      @batch_template.delete(:name)
      @batch_template.delete(:topic)
      @batch_template.delete(:dataType)
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- dataType (json field in request body) is a required field\n- name (json field in request body) is a required field\n- topic (json field in request body) is a required field"
    end

    it 'Missing Tenant ID' do
      response = hri_post_batch(nil, @batch_template.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json,{ 'Authorization' => "No auth token" })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_post_batch(TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - No Roles' do
      response = hri_post_batch(TENANT_ID_WITH_NO_ROLES, @batch_template.to_json)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to create a batch'
    end

    it 'Unauthorized - Incorrect Roles' do
      response = hri_post_batch(TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_post_batch(TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 401
    end

  end




  context 'GET /tenants/{tenant_id}/batches' do

    it 'Success Without Status' do
      response = hri_get_batches(AUTHORIZED_TENANT_ID, 'size=1000')
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(batch['id']).to_not be_nil
        expect(batch['name']).to_not be_nil
        expect(batch['topic']).to_not be_nil
        expect(batch['dataType']).to_not be_nil
        expect(%w(started completed failed terminated sendCompleted)).to include batch['status']
        expect(batch['startDate']).to_not be_nil
      end
    end

    it 'Success With Status' do
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "status=#{STATUS}&size=1000")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(%w(started completed failed terminated sendCompleted)).to include batch['status']
      end
    end

    it 'Success With Name' do
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "name=#{@batch_name}&size=1000")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
    end

    it 'Success With Greater Than Date' do
      greater_than_date = Date.today - 365
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "gteDate=#{greater_than_date}&size=1000")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(DateTime.strptime(batch['startDate'], '%Y-%m-%dT%H:%M:%S%Z')).to be > greater_than_date
      end
    end

    it 'Success With Less Than Date' do
      less_than_date = Date.today + 1
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "lteDate=#{less_than_date}&size=1000")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(DateTime.strptime(batch['startDate'], '%Y-%m-%dT%H:%M:%S%Z')).to be < less_than_date
      end
    end

    it 'Name Not Found' do
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "name=#{INVALID_ID}")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be false
    end

    it 'Status Not Found' do
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "status=#{INVALID_ID}")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be false
    end

    it 'Greater Than Date With No Results' do
      greater_than_date = Date.today + 10000
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "gteDate=#{greater_than_date}")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be false
    end

    it 'Less Than Date With No Results' do
      less_than_date = Date.today - 5000
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, "lteDate=#{less_than_date}")
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be false
    end


    it 'Query Parameter With Restricted Characters' do
      response = hri_get_batches(AUTHORIZED_TENANT_ID, 'status="[{started}]"')
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- status (request query parameter) must not contain the following characters: \"=<>[]{}"
    end

    it 'Missing Tenant ID' do
      response = hri_get_batches(nil, 'size=1000')
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Tenant ID can not view batches created with a different Tenant ID' do
      #Create Batch
      @batch_prefix = 'batch'
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        topic: BATCH_INPUT_TOPIC,
        dataType: "#{DATA_TYPE}",
        invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Verify Batch Not Visible to Different Integrator ID
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, 'size=1000')
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      unless parsed_response['results'].empty?
        parsed_response['results'].each do |batch|
          raise "Batch ID #{@new_batch_id} found with different Integrator ID!" if batch['id'] == @new_batch_id
        end
      end

      #Verify Batch Visible To Consumer Role
      @batch_found = false
      response = hri_get_batches(TENANT_WITH_INTEGRATOR_CONSUMER_ROLE, 'size=1000')
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      parsed_response['results'].each do |batch|
        if batch['id'] == @new_batch_id
          @batch_found = true
        end
      end
      expect(@batch_found).to be false
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_get_batches(AUTHORIZED_TENANT_ID, nil,{'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_get_batches(TENANT_ID, nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - No Roles' do
      response = hri_get_batches(TENANT_ID_WITH_NO_ROLES, nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'The access token must have one of these scopes: hri_consumer, hri_data_integrator'
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_get_batches(TENANT_ID, nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

  end

  context 'GET /tenants/{tenantId}/batches/{batchId}' do

    it 'Success With Consumer Role Only' do
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['id']).to eq @batch_id
      expect(parsed_response['name']).to eql @batch_name
      expect(parsed_response['dataType']).to eql DATA_TYPE
      expect(parsed_response['topic']).to eql BATCH_INPUT_TOPIC
      expect(parsed_response['invalidThreshold']).to eql "#{INVALIDTHRESHOLD}".to_i
      expect(parsed_response['metadata']['compression']).to eql 'gzip'
      expect(parsed_response['metadata']['finalRecordCount']).to eql 20

    end

    it 'Success With Integrator Role Only' do
      #Create Batch
      @batch_prefix = 'batch'
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        topic: BATCH_INPUT_TOPIC,
        dataType: "#{DATA_TYPE}",
        invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Get Batch
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @new_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['id']).to eq @new_batch_id
      expect(parsed_response['name']).to eql @batch_name
      expect(parsed_response['status']).to eql STATUS
      expect(parsed_response['dataType']).to eql DATA_TYPE
      expect(parsed_response['topic']).to eql BATCH_INPUT_TOPIC
      expect(parsed_response['invalidThreshold']).to eql "#{INVALIDTHRESHOLD}".to_i
      expect(parsed_response['metadata']['compression']).to eql 'gzip'
      expect(parsed_response['metadata']['finalRecordCount']).to eql 20
    end

    it 'Missing Tenant ID' do
      response = hri_get_batch(nil, @batch_id)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Batch ID Not Found' do
      response = hri_get_batch(AUTHORIZED_TENANT_ID, INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The document for tenantId: #{AUTHORIZED_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Integrator ID can not view a batch created with a different Integrator ID' do
      #Create Batch
      @batch_prefix = 'batch'
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        topic: BATCH_INPUT_TOPIC,
        dataType: "#{DATA_TYPE}",
        invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Verify Batch Not Visible to Different Integrator ID
      response = hri_get_batch(TENANT_ID, @new_batch_id)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @batch_id,{'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_get_batch(TENANT_ID, @batch_id)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - No Roles' do
      response = hri_get_batch(TENANT_ID_WITH_NO_ROLES, @batch_id)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'The access token must have one of these scopes: hri_consumer, hri_data_integrator'
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_get_batch(TENANT_ID, @batch_id)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Missing Batch ID With Ending Forward Slash' do
      response = hri_custom_request("/tenants/#{AUTHORIZED_TENANT_ID}/batches//", nil, 'GET')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/sendComplete' do

    before(:all) do
      @expected_record_count = {
        expectedRecordCount: 1,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 15
        }
      }

      @terminate_metadata = {
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      @batch_prefix = 'batch'
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        topic: BATCH_INPUT_TOPIC,
        dataType: "#{DATA_TYPE}",
        invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }
    end

    it 'Success' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Set Batch Complete
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 200

      #Verify Batch Complete
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      expect(parsed_response['endDate']).to_not be_nil

    end

    it 'Success with recordCount' do
      record_count = {
        expectedRecordCount: 200,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Set Batch Complete
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', record_count)
      expect(response.code).to eq 200

      #Verify Batch Complete
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      expect(parsed_response['endDate']).to_not be_nil

    end

    it 'Invalid Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, INVALID_ID, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{AUTHORIZED_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing Record Count' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'sendComplete', nil)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- expectedRecordCount (json field in request body) must be present if recordCount (json field in request body) is not present\n- recordCount (json field in request body) must be present if expectedRecordCount (json field in request body) is not present"
    end

    it 'Invalid Record Count' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'sendComplete', {expectedRecordCount: "1"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"expectedRecordCount\": expected type int, but received type string"
    end

    it 'Missing Tenant ID' do
      response = hri_put_batch(nil, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Missing Batch ID and Record Count' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'sendComplete', nil)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- expectedRecordCount (json field in request body) must be present if recordCount (json field in request body) is not present\n- id (url path parameter) is a required field\n- recordCount (json field in request body) must be present if expectedRecordCount (json field in request body) is not present"
    end

    it 'Conflict: Batch with a status other than started' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Terminated Status
      updated_status = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 201

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      expect(parsed_response['endDate']).to_not be_nil

      #Attempt to complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'terminated' state"

    end

    it 'Conflict: Batch that already has a completed status' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Completed Status
      updated_status = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'

      #Attempt to complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'completed' state"

    end

    it 'Integrator ID can not update batches created with a different Integrator ID' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Verify Batch Not Updated With Different Integrator ID
      response = hri_put_batch(TEST_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "Unauthorized tenant access. Tenant '#{TEST_TENANT_ID}' is not included in the authorized roles:tenant_#{TEST_TENANT_ID}."
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - No Roles' do
      response = hri_put_batch(TENANT_ID_WITH_NO_ROLES, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{TENANT_ID_WITH_NO_ROLES} with document (batch) ID: #{@batch_id} was not found"
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end
  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/terminate' do

    before(:all) do
      @batch_prefix = 'batch'
      @terminate_batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @terminate_batch_name,
        topic: BATCH_INPUT_TOPIC,
        dataType: "#{DATA_TYPE}",
        invalidThreshold: "#{INVALIDTHRESHOLD}".to_i,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      @expected_record_count = {
        expectedRecordCount: 1,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 15
        }
      }

      @terminate_metadata = {
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }
    end

    it 'Success' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Terminate Batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 200

      #Verify Batch Terminated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      expect(parsed_response['endDate']).to_not be_nil

    end

    it 'Invalid Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, INVALID_ID, 'terminate', nil)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{AUTHORIZED_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing Tenant ID' do
      response = hri_put_batch(nil, @batch_id, 'terminate', @expected_record_count)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'terminate', @expected_record_count)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Conflict: Batch with a status other than started' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Completed Status
      updated_status = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 201

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Attempt to terminate batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', nil)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'completed' state"

    end

    it 'Conflict: Batch that already has a terminated status' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Completed Status
      updated_status = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', nil)

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Attempt to terminate batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', nil)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'terminated' state"

    end

    it 'Integrator ID can not update batches created with a different Integrator ID' do
      #Create Batch
      response = hri_post_batch(TENANT_WITH_DATA_INTEGRATOR_ROLE, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")


      #Verify Batch Not Updated With Different Integrator ID
      response = hri_put_batch(TEST_TENANT_ID, @terminate_batch_id, 'terminate', nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "Unauthorized tenant access. Tenant '#{TEST_TENANT_ID}' is not included in the authorized roles:tenant_#{TEST_TENANT_ID}."

    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'terminate', nil,{'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'terminate', nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - No Roles' do
      response = hri_put_batch(TENANT_ID_WITH_NO_ROLES, @batch_id, 'terminate', nil)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{TENANT_ID_WITH_NO_ROLES} with document (batch) ID: #{@batch_id} was not found"
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      response = hri_put_batch(TENANT_WITH_DATA_CONSUMER_ROLE, @batch_id, 'terminate', nil)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{TENANT_WITH_DATA_CONSUMER_ROLE} with document (batch) ID: #{@batch_id} was not found"
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'terminate', nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

  end



end

