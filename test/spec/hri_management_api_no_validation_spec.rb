# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'
require 'mongo'

describe 'HRI Management API Without Validation' do

  INVALID_ID = 'INVALID'
  TENANT_ID = 'test0211'
  INTEGRATOR_ID = 'claims'
  TEST_TENANT_ID = "rspec-#{'2311'.delete('.')}-test-tenant".downcase
  TEST_INTEGRATOR_ID = "rspec-#{'2311'.delete('.')}-test-integrator".downcase
  DATA_TYPE = 'rspec-batch'
  STATUS = 'started'
  BATCH_INPUT_TOPIC = "ingest.#{TENANT_ID}.#{INTEGRATOR_ID}.in"
  KAFKA_TIMEOUT = 60


  def initialize(mongodb_credentials = {})
    @headers = { 'Content-Type': 'application/json' }
    client = Mongo::Client.new('mongodb://hi-dp-tst-eastus-cosmos-mongo-api-hri:Jl6rN2wUFpROlr4Cxse61ET51TB1qwZTZXfD1IwotXQKUBUaEGjBXr8DqKAKonhBkhwSxdLIkJitZUE9X2liSg==@hi-dp-tst-eastus-cosmos-mongo-api-hri.mongo.cosmos.azure.com:10255/?ssl=true&replicaSet=globaldb&retrywrites=false&maxIdleTimeMS=120000&appName=@hi-dp-tst-eastus-cosmos-mongo-api-hri@' , :database => 'HRI-DEV')
    db = client.database
    collection = client[:'HRI-Mgmt']
    #puts collection.find( { tenantid: 'q-batches' } ).first
  end

  def get_client_id_and_secret()
    return ['c33ac4da-21c6-426b-abcc-27e24ff1ccf9', 'GxF8Q~XfZyLRQBZ4mjwgEogVWwGjtzJh7ZPzgagw']
  end

  def get_access_token()
    credentials = get_client_id_and_secret()
    response = @request_helper.rest_post("https://login.microsoftonline.com/ceaa63aa-5d5c-4c7d-94b0-02f9a3ab6a8c/oauth2/v2.0/token",{'grant_type' => 'client_credentials','scope' => 'c33ac4da-21c6-426b-abcc-27e24ff1ccf9/.default', 'client_secret' => 'GxF8Q~XfZyLRQBZ4mjwgEogVWwGjtzJh7ZPzgagw', 'client_id' => 'c33ac4da-21c6-426b-abcc-27e24ff1ccf9'}, {'Content-Type' => 'application/x-www-form-urlencoded', 'Accept' => 'application/json', 'Authorization' => "Basic #{Base64.encode64("#{credentials[0]}:#{credentials[1]}").delete("\n")}" })
    raise 'App ID token request failed' unless response.code == 200
    #puts "This is the generated token ==============================>  "
    #puts JSON.parse(response.body)['access_token']
    JSON.parse(response.body)['access_token']
  end

  def hri_post_tenant(tenant_id, request_body = nil, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_post(url, request_body, headers)
  end

  def hri_post_tenant_stream(tenant_id, integrator_id, request_body, override_headers = {}, delete_auth = false)
    url = "#{@hri_base_url}/tenants/#{tenant_id}/streams/#{integrator_id}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    headers.delete('Authorization') if delete_auth
    rest_post(url, request_body, headers, {})
  end

  def hri_delete_tenant(tenant_id, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_delete(url, nil, headers, {})
  end

  def hri_get_tenant_streams(tenant_id, override_headers = {}, delete_auth = false)
    url = "#{@hri_base_url}/tenants/#{tenant_id}/streams"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    headers.delete('Authorization') if delete_auth
    rest_get(url, headers)
  end


  def hri_get_tenant(tenant_id, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_get_tenants(override_headers = {})
    url = "#{@hri_base_url}/tenants"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_delete_tenant_stream(tenant_id, integrator_id, override_headers = {}, delete_auth = false)
    url = "#{@hri_base_url}/tenants/#{tenant_id}/streams/#{integrator_id}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    headers.delete('Authorization') if delete_auth
    rest_delete(url, nil, headers, {})
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
    @az_token = get_access_token()
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
    #@elastic = HRITestHelpers::ElasticHelper.new({url: ENV['ELASTIC_URL'], username: ENV['ELASTIC_USERNAME'], password: ENV['ELASTIC_PASSWORD']})
    @iam_token = get_access_token()
    #@azure_token = HRITestHelpers::AppIDHelper.new(get_access_token(@appid_url, @appid_tenant))
    #@mgmt_api_helper = HRITestHelpers::MgmtAPIHelper.new(@hri_base_url, @iam_token)
    #@hri_deploy_helper = HRIDeployHelper.new
    #@event_streams_api_helper = HRITestHelpers::EventStreamsAPIHelper.new(ENV['ES_ADMIN_URL'], ENV['CLOUD_API_KEY'])
    #@app_id_helper = HRITestHelpers::initialize.new()
    @start_date = DateTime.now

    @exe_path = File.absolute_path(File.join(File.dirname(__FILE__), "../../src/hri"))
    @config_path = File.absolute_path(File.join(File.dirname(__FILE__), "test_config"))
    @log_path = File.absolute_path(File.join(File.dirname(__FILE__), "../logs"))
    Dir.mkdir(@log_path) if !Dir.exists?(@log_path)


    #@hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path, 'no-validation-1-')
    #response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
    #unless response.code == 200
    #raise "Health check failed: #{response.body}"
    #end

    #Initialize Kafka Consumer
    @kafka = Kafka.new(ENV['KAFKA_BROKERS'], sasl_plain_username: 'token', sasl_plain_password: ENV['KAFKA_PASSWORD'], ssl_ca_certs_from_system: true)
    @kafka_consumer = @kafka.consumer(group_id: 'rspec-mgmt-api-consumer')
    @kafka_consumer.subscribe("ingest.#{TENANT_ID}.#{INTEGRATOR_ID}.notification")

    #Get AppId Access Tokens
    #@token_invalid_tenant = @app_id_helper.get_access_token('hri_integration_tenant_test_invalid', 'tenant_test_invalid')
    #@token_no_roles = @app_id_helper.get_access_token('hri_integration_tenant_test', 'tenant_test')
    #@token_integrator_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_integrator', 'tenant_test hri_data_integrator')
    #@token_consumer_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_consumer', 'tenant_test hri_consumer')
    #@token_all_roles = @app_id_helper.get_access_token(@appid_url, @appid_tenant)
    #puts token_all_roles
    #@token_invalid_audience = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer', ENV['APPID_TENANT'])
  end


  context 'POST /tenants/{tenant_id}' do

    it 'Success' do
      response = hri_post_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql TEST_TENANT_ID
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

  context 'GET /tenants' do

    it 'Success' do
      response = hri_get_tenants
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].to_s).to include TENANT_ID
    end

    it 'Unauthorized' do
      response = hri_get_tenants({ 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

  end

  context 'GET /tenants/{tenant_id}' do

    it 'Success' do
      response = hri_get_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql "#{TEST_TENANT_ID}-batches"
      expect(parsed_response['health']).to eql 'green'
      expect(parsed_response['status']).to eql 'open'
    end

    it 'Invalid Tenant ID' do
      response = hri_get_tenant(INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Tenant: #{INVALID_ID} not found: [404]"
    end

    it 'Unauthorized' do
      response = hri_get_tenant(INVALID_ID, {'Authorization': nil})
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

      #Timeout.timeout(30, nil, 'Kafka topics not created after 30 seconds') do
      #loop do
      #topics = @event_streams_helper.get_topics
      #break if (topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.in") && topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.notification"))
      #end
      #end
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

    #it 'Invalid cleanupPolicy' do
    #@stream_info[:cleanupPolicy] = 12345
    #response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
    #expect(response.code).to eq 400
    #parsed_response = JSON.parse(response.body)
    #expect(parsed_response['errorDescription']).to eql "invalid request param \"cleanupPolicy\": expected type string, but received type number"
    #end

    #it 'cleanupPolicy must be "compact" or "delete"' do
    #@stream_info[:cleanupPolicy] = "invalid"
    #response = hri_post_tenant_stream(TEST_TENANT_ID, 'test', @stream_info.to_json)
    #expect(response.code).to eq 400
    #parsed_response = JSON.parse(response.body)
    #expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- cleanupPolicy (json field in request body) must be one of [delete compact]"
    #end

    it 'Invalid cleanupPolicy and missing numPartitions' do
      #@stream_info[:cleanupPolicy] = INVALID_ID
      @stream_info.delete(:numPartitions)
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- numPartitions (json field in request body) is a required field"
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


  context 'GET /tenants/{tenant_id}/streams' do

    it 'Success' do
      response = hri_get_tenant_streams(TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'][0]['id']).to eql INTEGRATOR_ID
    end

    #it 'Success With Invalid Topic Only' do
    #invalid_topic = "ingest.#{TENANT_ID}.#{TEST_INTEGRATOR_ID}.invalid"
    #@event_streams_helper.create_topic(invalid_topic, 1)
    #Timeout.timeout(30, nil, "Timed out waiting for the '#{invalid_topic}' topic to be created") do
    #loop do
    #break if @event_streams_helper.get_topics.include?(invalid_topic)
    #end
    #end

    #response = hri_get_tenant_streams(TENANT_ID)
    #expect(response.code).to eq 200
    #parsed_response = JSON.parse(response.body)
    #stream_found = false
    #parsed_response['results'].each do |integrator|
    #stream_found = true if integrator['id'] == TEST_INTEGRATOR_ID
    #end
    #raise "Tenant Stream Not Found: #{TEST_INTEGRATOR_ID}" unless stream_found

    #@event_streams_helper.delete_topic(invalid_topic)
    #Timeout.timeout(30, nil, "Timed out waiting for the '#{invalid_topic}' topic to be deleted") do
    #loop do
    #break unless @event_streams_helper.get_topics.include?(invalid_topic)
    #end
    #end
    #end

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


  context 'DELETE /tenants/{tenant_id}/streams/{integrator_id}' do

    it 'Success' do
      #Delete Stream and Verify Deletion
      Timeout.timeout(20, nil, 'Kafka topics not deleted after 20 seconds') do
        loop do
          response = hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID)
          break if response.code == 200

          response = hri_get_tenant_streams(TEST_TENANT_ID)
          expect(response.code).to eq 200
          parsed_response = JSON.parse(response.body)
          break if parsed_response['results'] == []
          sleep 1
        end
      end
    end

    it 'Invalid Stream' do
      response = hri_delete_tenant_stream(INVALID_ID, TEST_INTEGRATOR_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unable to delete topic \"ingest.INVALID.rspec-2311-test-integrator.in\": Broker: Unknown topic or partition\nUnable to delete topic \"ingest.INVALID.rspec-2311-test-integrator.notification\": Broker: Unknown topic or partition"
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
      response = hri_delete_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 200

      #Verify Tenant Deleted
      response = hri_get_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 404
    end

    it 'Invalid Tenant ID' do
      response = hri_delete_tenant(INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Could not delete tenant [#{INVALID_ID}-batches]: [404]"
    end

    it 'Unauthorized' do
      response = hri_delete_tenant(TEST_TENANT_ID, { 'Authorization': nil })
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


end