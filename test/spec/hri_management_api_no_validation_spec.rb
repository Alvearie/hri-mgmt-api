# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

describe 'HRI Management API Without Validation' do

  INVALID_ID = 'INVALID'
  EXISTING_TENANT_ID = ENV['EXISTING_TENANT_ID']
  EXISTING_INTEGRATOR_ID = 'claims'
  TEST_TENANT_ID = "rspec-#{ENV['BRANCH_NAME'].delete('.')}-test-tenant".downcase
  TEST_INTEGRATOR_ID = "rspec-#{ENV['BRANCH_NAME'].delete('.')}-test-integrator".downcase
  DATA_TYPE = 'rspec-batch'
  STATUS = 'started'
  BATCH_INPUT_TOPIC = "ingest.#{EXISTING_TENANT_ID}.#{EXISTING_INTEGRATOR_ID}.in"
  KAFKA_TIMEOUT = 60

  before(:all) do
    @hri_base_url = ENV['HRI_URL']
    @request_helper = HRITestHelpers::RequestHelper.new
    @elastic = HRITestHelpers::ElasticHelper.new({url: ENV['ELASTIC_URL'], username: ENV['ELASTIC_USERNAME'], password: ENV['ELASTIC_PASSWORD']})
    @iam_token = HRITestHelpers::IAMHelper.new(ENV['IAM_CLOUD_URL']).get_access_token(ENV['CLOUD_API_KEY'])
    @mgmt_api_helper = HRITestHelpers::MgmtAPIHelper.new(@hri_base_url, @iam_token)
    @hri_deploy_helper = HRIDeployHelper.new
    @event_streams_helper = HRITestHelpers::EventStreamsHelper.new
    @app_id_helper = HRITestHelpers::AppIDHelper.new(ENV['APPID_URL'], ENV['APPID_TENANT'], @iam_token, ENV['JWT_AUDIENCE_ID'])
    @start_date = DateTime.now

    @exe_path = File.absolute_path(File.join(File.dirname(__FILE__), "../../src/hri"))
    @config_path = File.absolute_path(File.join(File.dirname(__FILE__), "test_config"))
    @log_path = File.absolute_path(File.join(File.dirname(__FILE__), "/"))

    @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
    response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
    unless response.code == 200
      raise "Health check failed: #{response.body}"
    end

    #Initialize Kafka Consumer
    @kafka = Kafka.new(ENV['KAFKA_BROKERS'], sasl_plain_username: 'token', sasl_plain_password: ENV['KAFKA_PASSWORD'], ssl_ca_certs_from_system: true)
    @kafka_consumer = @kafka.consumer(group_id: 'rspec-mgmt-api-consumer')
    @kafka_consumer.subscribe("ingest.#{EXISTING_TENANT_ID}.#{EXISTING_INTEGRATOR_ID}.notification")

    #Create Batch
    @batch_prefix = "rspec-#{ENV['BRANCH_NAME'].delete('.')}"
    @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
    create_batch = {
      name: @batch_name,
      status: STATUS,
      recordCount: 1,
      dataType: DATA_TYPE,
      topic: BATCH_INPUT_TOPIC,
      startDate: @start_date,
      metadata: {
        rspec1: 'test1',
        rspec2: 'test2',
        rspec3: {
          rspec3A: 'test3A',
          rspec3B: 'test3B'
        }
      }
    }.to_json
    response = @elastic.es_create_batch(EXISTING_TENANT_ID, create_batch)
    expect(response.code).to eq 201
    parsed_response = JSON.parse(response.body)
    @batch_id = parsed_response['_id']
    Logger.new(STDOUT).info("New Batch Created With ID: #{@batch_id}")

    #Get AppId Access Tokens
    @token_invalid_tenant = @app_id_helper.get_access_token('hri_integration_tenant_test_invalid', 'tenant_test_invalid', ENV['JWT_AUDIENCE_ID'])
    @token_no_roles = @app_id_helper.get_access_token('hri_integration_tenant_test', 'tenant_test', ENV['JWT_AUDIENCE_ID'])
    @token_integrator_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_integrator', 'tenant_test hri_data_integrator', ENV['JWT_AUDIENCE_ID'])
    @token_consumer_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_consumer', 'tenant_test hri_consumer', ENV['JWT_AUDIENCE_ID'])
    @token_all_roles = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer', ENV['JWT_AUDIENCE_ID'])
    @token_invalid_audience = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer', ENV['APPID_TENANT'])
  end

  after(:all) do
    File.delete("#{@log_path}/output.txt") if File.exists?("#{@log_path}/output.txt")
    File.delete("#{@log_path}/error.txt") if File.exists?("#{@log_path}/error.txt")

    processes = `lsof -iTCP:1323 -sTCP:LISTEN`
    unless processes == ''
      process_id = processes.split("\n").select { |s| s.start_with?('hri') }[0].split(' ')[1].to_i
      `kill #{process_id}` unless process_id.nil?
    end

    #Delete Batches
    response = @elastic.es_delete_by_query(EXISTING_TENANT_ID, "name:rspec-#{ENV['BRANCH_NAME']}*")
    response.nil? ? (raise 'Elastic batch delete did not return a response') : (expect(response.code).to eq 200)
    Logger.new(STDOUT).info("Delete test batches by query response #{response.body}")

    @kafka_consumer.stop
  end

  context 'POST /tenants/{tenant_id}' do

    it 'Success' do
      response = @mgmt_api_helper.hri_post_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql TEST_TENANT_ID
    end

    it 'Tenant Already Exists' do
      response = @mgmt_api_helper.hri_post_tenant(EXISTING_TENANT_ID)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "Unable to create new tenant [#{EXISTING_TENANT_ID}]: [400] resource_already_exists_exception"
    end

    it 'Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_post_tenant(INVALID_ID)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"
    end

    it 'Unauthorized' do
      response = @mgmt_api_helper.hri_post_tenant(TEST_TENANT_ID, nil, { 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'elastic IAM authentication returned 401 : 401 Unauthorized'
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_post_tenant(nil)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Tenant ID With Ending Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request('/tenants//', nil, {}, 'POST')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Tenant ID With No Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request('/tenants', nil, {}, 'POST')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Invalid Elastic CRN' do
      process_id = `lsof -iTCP:1323 -sTCP:LISTEN`.split("\n").select { |s| s.start_with?('hri') }[0].split(' ')[1].to_i
      `kill #{process_id}` unless process_id.nil?
      temp = ENV['ELASTIC_CRN']
      ENV['ELASTIC_CRN'] = 'INVALID'
      @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
      response = @mgmt_api_helper.hri_post_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 500
      ENV['ELASTIC_CRN'] = temp
      process_id = `lsof -iTCP:1323 -sTCP:LISTEN`.split("\n").select { |s| s.start_with?('hri') }[0].split(' ')[1].to_i
      `kill #{process_id}` unless process_id.nil?
      @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path)
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
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 201

      #Verify Stream Creation
      response = @mgmt_api_helper.hri_get_tenant_streams(TEST_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'][0]['id']).to eql TEST_INTEGRATOR_ID

      Timeout.timeout(30, nil, 'Kafka topics not created after 30 seconds') do
        loop do
          topics = @event_streams_helper.get_topics
          break if (topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.in") && topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.notification"))
        end
      end
    end

    it 'Stream Already Exists' do
      response = @mgmt_api_helper.hri_post_tenant_stream(EXISTING_TENANT_ID, EXISTING_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "topic 'ingest.#{EXISTING_TENANT_ID}.#{EXISTING_INTEGRATOR_ID}.in' already exists"
    end

    it 'Missing numPartitions' do
      @stream_info.delete(:numPartitions)
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- numPartitions (json field in request body) is a required field"
    end

    it 'Invalid numPartitions' do
      @stream_info[:numPartitions] = '1'
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"numPartitions\": expected type int64, but received type string"
    end

    it 'Missing retentionMs' do
      @stream_info.delete(:retentionMs)
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- retentionMs (json field in request body) is a required field"
    end

    it 'Invalid retentionMs' do
      @stream_info[:retentionMs] = '3600000'
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"retentionMs\": expected type int, but received type string"
    end

    it 'Invalid Stream Name' do
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, ".#{TEST_INTEGRATOR_ID}.#{TEST_INTEGRATOR_ID}", @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'"
    end

    it 'Invalid cleanupPolicy' do
      @stream_info[:cleanupPolicy] = 12345
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"cleanupPolicy\": expected type string, but received type number"
    end

    it 'cleanupPolicy must be "compact" or "delete"' do
      @stream_info[:cleanupPolicy] = "invalid"
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, 'test', @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- cleanupPolicy (json field in request body) must be one of [delete compact]"
    end

    it 'Invalid cleanupPolicy and missing numPartitions' do
      @stream_info[:cleanupPolicy] = INVALID_ID
      @stream_info.delete(:numPartitions)
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- cleanupPolicy (json field in request body) must be one of [delete compact]\n- numPartitions (json field in request body) is a required field"
    end

    it 'Missing Authorization' do
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json, {'Authorization' => nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "missing header 'Authorization'"
    end

    it 'Invalid Authorization' do
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json, {'Authorization' => 'Bearer Invalid'})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Unauthorized to manage resource'
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_post_tenant_stream(nil, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 400
    end

    it 'Missing Stream ID' do
      response = @mgmt_api_helper.hri_post_tenant_stream(TEST_TENANT_ID, nil, @stream_info.to_json)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Stream ID With Ending Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams//", @stream_info.to_json, {}, 'POST')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Stream ID With No Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams", @stream_info.to_json, {}, 'POST')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'DELETE /tenants/{tenant_id}/streams/{integrator_id}' do

    it 'Success' do
      #Delete Stream and Verify Deletion
      Timeout.timeout(20, nil, 'Kafka topics not deleted after 20 seconds') do
        loop do
          response = @mgmt_api_helper.hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID)
          break if response.code == 200

          response = @mgmt_api_helper.hri_get_tenant_streams(TEST_TENANT_ID)
          expect(response.code).to eq 200
          parsed_response = JSON.parse(response.body)
          break if parsed_response['results'] == []
          sleep 1
        end
      end
    end

    it 'Invalid Stream' do
      response = @mgmt_api_helper.hri_delete_tenant_stream(INVALID_ID, TEST_INTEGRATOR_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unable to delete topic \"ingest.#{INVALID_ID}.#{TEST_INTEGRATOR_ID}.in\": topic 'ingest.#{INVALID_ID}.#{TEST_INTEGRATOR_ID}.in' does not exist\nUnable to delete topic \"ingest.#{INVALID_ID}.#{TEST_INTEGRATOR_ID}.notification\": topic 'ingest.#{INVALID_ID}.#{TEST_INTEGRATOR_ID}.notification' does not exist"
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_delete_tenant_stream(nil, EXISTING_INTEGRATOR_ID)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Authorization' do
      response = @mgmt_api_helper.hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, {'Authorization' => nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unable to delete topic \"ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.in\": missing header 'Authorization'\nUnable to delete topic \"ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.notification\": missing header 'Authorization'"
    end

    it 'Invalid Authorization' do
      response = @mgmt_api_helper.hri_delete_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, {'Authorization' => 'Bearer Invalid'})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unable to delete topic \"ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.in\": Unauthorized to manage resource\nUnable to delete topic \"ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.notification\": Unauthorized to manage resource"
    end

    it 'Missing Stream ID' do
      response = @mgmt_api_helper.hri_delete_tenant_stream(TEST_TENANT_ID, nil)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Stream ID With Ending Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams//", nil, {}, 'DELETE')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Stream ID With No Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request("/tenants/#{TEST_TENANT_ID}/streams", nil, {}, 'DELETE')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'DELETE /tenants/{tenant_id}' do

    it 'Success' do
      #Delete Tenant
      response = @mgmt_api_helper.hri_delete_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 200

      #Verify Tenant Deleted
      response = @mgmt_api_helper.hri_get_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 404
    end

    it 'Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_delete_tenant(INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Could not delete tenant [#{INVALID_ID}]: [404] index_not_found_exception: no such index [#{INVALID_ID}-batches]"
    end

    it 'Unauthorized' do
      response = @mgmt_api_helper.hri_delete_tenant(TEST_TENANT_ID, { 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'elastic IAM authentication returned 401 : 401 Unauthorized'
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_delete_tenant(nil)
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

    it 'Missing Tenant ID With Ending Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request('/tenants//', nil, {}, 'DELETE')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

    it 'Missing Tenant ID With No Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request('/tenants', nil, {}, 'DELETE')
      expect(response.code).to eq 405
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Method Not Allowed'
    end

  end

  context 'GET /tenants/{tenant_id}/streams' do

    it 'Success' do
      response = @mgmt_api_helper.hri_get_tenant_streams(EXISTING_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'][0]['id']).to eql EXISTING_INTEGRATOR_ID
    end

    it 'Success With Invalid Topic Only' do
      invalid_topic = "ingest.#{EXISTING_TENANT_ID}.#{TEST_INTEGRATOR_ID}.invalid"
      @event_streams_helper.create_topic(invalid_topic, 1)
      response = @mgmt_api_helper.hri_get_tenant_streams(EXISTING_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      stream_found = false
      parsed_response['results'].each do |integrator|
        stream_found = true if integrator['id'] == TEST_INTEGRATOR_ID
      end
      raise "Tenant Stream Not Found: #{TEST_INTEGRATOR_ID}" unless stream_found

      Timeout.timeout(15, nil, "Timed out waiting for the '#{invalid_topic}' topic to be deleted") do
        loop do
          break if @event_streams_helper.get_topics.include?(invalid_topic)
        end
        loop do
          @event_streams_helper.delete_topic(invalid_topic)
          break unless @event_streams_helper.get_topics.include?(invalid_topic)
        end
      end
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_get_tenant_streams(nil)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Authorization' do
      response = @mgmt_api_helper.hri_get_tenant_streams(EXISTING_TENANT_ID, {'Authorization' => nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "missing header 'Authorization'"
    end

    it 'Invalid Authorization' do
      response = @mgmt_api_helper.hri_get_tenant_streams(EXISTING_TENANT_ID, {'Authorization' => 'Bearer Invalid'})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Unauthorized to manage resource'
    end

  end

  context 'GET /tenants' do

    it 'Success' do
      response = @mgmt_api_helper.hri_get_tenants
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].to_s).to include EXISTING_TENANT_ID
    end

    it 'Unauthorized' do
      response = @mgmt_api_helper.hri_get_tenants({ 'Authorization': nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'elastic IAM authentication returned 401 : 401 Unauthorized'
    end

  end

  context 'GET /tenants/{tenant_id}' do

    it 'Success' do
      response = @mgmt_api_helper.hri_get_tenant(EXISTING_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['index']).to eql "#{EXISTING_TENANT_ID}-batches"
      expect(parsed_response['health']).to eql 'green'
      expect(parsed_response['uuid']).to_not be_nil
    end

    it 'Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_get_tenant(INVALID_ID)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Tenant: #{INVALID_ID} not found: [404] index_not_found_exception: no such index [#{INVALID_ID}-batches]"
    end

    it 'Unauthorized' do
      response = @mgmt_api_helper.hri_get_tenant(INVALID_ID, {'Authorization': nil})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'elastic IAM authentication returned 401 : 401 Unauthorized'
    end

    it 'Ending Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request('/tenants//', nil, {}, 'GET')
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['message']).to eql 'Not Found'
    end

  end

  context 'POST /tenants/{tenant_id}/batches' do

    before(:each) do
      @batch_name = "#{@batch_prefix}-テスト-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        dataType: "#{DATA_TYPE}-テスト",
        topic: BATCH_INPUT_TOPIC,
        startDate: Date.today,
        endDate: Date.today + 1,
        metadata: {
          rspec1: 'test1',
          rspec2: 'test2',
          rspec3: {
            rspec3A: 'test3A',
            rspec3B: 'テスト'
          }
        }
      }
    end

    it 'Successful Batch Creation' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 201
      expect(response.headers[:'content_security_policy']).to eql "default-src 'none'; script-src 'none'; connect-src 'self'; img-src 'self'; style-src 'self';"
      expect(response.headers[:'x_content_type_options']).to eql 'nosniff'
      expect(response.headers[:'x_xss_protection']).to eql '1'
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Verify Batch in Elastic
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @new_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_index']).to eql "#{EXISTING_TENANT_ID}-batches"
      expect(parsed_response['_id']).to eql @new_batch_id
      expect(parsed_response['found']).to be true
      expect(parsed_response['_source']['name']).to eql @batch_name
      expect(parsed_response['_source']['topic']).to eql BATCH_INPUT_TOPIC
      expect(parsed_response['_source']['dataType']).to eql "#{DATA_TYPE}-テスト"
      expect(parsed_response['_source']['metadata']['rspec1']).to eql 'test1'
      expect(parsed_response['_source']['metadata']['rspec2']).to eql 'test2'
      expect(parsed_response['_source']['metadata']['rspec3']['rspec3A']).to eql 'test3A'
      expect(parsed_response['_source']['metadata']['rspec3']['rspec3B']).to eql 'テスト'
      expect(DateTime.parse(parsed_response['_source']['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@new_batch_id} and status: #{STATUS}")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @new_batch_id and parsed_message['status'] == STATUS
            @message_found = true
            expect(parsed_message['dataType']).to eql "#{DATA_TYPE}-テスト"
            expect(parsed_message['id']).to eql @new_batch_id
            expect(parsed_message['name']).to eql @batch_name
            expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
            expect(parsed_message['status']).to eql STATUS
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(parsed_message['metadata']['rspec1']).to eql 'test1'
            expect(parsed_message['metadata']['rspec2']).to eql 'test2'
            expect(parsed_message['metadata']['rspec3']['rspec3A']).to eql 'test3A'
            expect(parsed_message['metadata']['rspec3']['rspec3B']).to eql 'テスト'
            break
          end
        end
        expect(@message_found).to be true
      end

      #Delete Batch
      response = @elastic.es_delete_batch(EXISTING_TENANT_ID, @new_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_id']).to eq @new_batch_id
      expect(parsed_response['result']).to eql 'deleted'
    end

    it 'should auto-delete a batch from Elastic if the batch was created with an invalid Kafka topic' do
      #Gather existing batches
      existing_batches = []
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, nil, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        existing_batches << batch['id'] unless batch['dataType'] == 'rspec-batch'
      end

      #Create Batch with Bad Topic
      @batch_template[:topic] = 'INVALID-TEST-TOPIC'
      @batch_template[:dataType] = 'rspec-invalid-batch'
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 500
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'kafka producer error: Broker: Unknown topic or partition'

      #Verify Batch Delete
      50.times do
        new_batches = []
        @batch_deleted = false
        response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, nil, { 'Authorization' => "Bearer #{@token_all_roles}" })
        expect(response.code).to eq 200
        parsed_response = JSON.parse(response.body)
        expect(parsed_response['total']).to be > 0
        parsed_response['results'].each do |batch|
          new_batches << batch['id'] unless batch['dataType'] == 'rspec-batch'
        end
        if existing_batches.sort == new_batches.sort
          @batch_deleted = true
          break
        end
      end

      expect(@batch_deleted).to be true
    end

    it 'Invalid Name' do
      @batch_template[:name] = 12345
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"name\": expected type string, but received type number"
    end

    it 'Invalid Topic' do
      @batch_template[:topic] = 12345
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"topic\": expected type string, but received type number"
    end

    it 'Invalid Data Type' do
      @batch_template[:dataType] = 12345
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"dataType\": expected type string, but received type number"
    end

    it 'Missing Name' do
      @batch_template.delete(:name)
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- name (json field in request body) is a required field"
    end

    it 'Missing Topic' do
      @batch_template.delete(:topic)
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- topic (json field in request body) is a required field"
    end

    it 'Missing Data Type' do
      @batch_template.delete(:dataType)
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- dataType (json field in request body) is a required field"
    end

    it 'Missing Name, Topic, and Data Type' do
      @batch_template.delete(:name)
      @batch_template.delete(:topic)
      @batch_template.delete(:dataType)
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- dataType (json field in request body) is a required field\n- name (json field in request body) is a required field\n- topic (json field in request body) is a required field"
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_post_batch(nil, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_all_roles}" })
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_invalid_tenant}" })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{EXISTING_TENANT_ID}' is not included in the authorized scopes: ."
    end

    it 'Unauthorized - No Roles' do
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_no_roles}" })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to create a batch'
    end

    it 'Unauthorized - Incorrect Roles' do
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_consumer_role_only}" })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to create a batch'
    end

    it 'Unauthorized - Invalid Audience' do
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, { 'Authorization' => "Bearer #{@token_invalid_audience}" })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Authorization token validation failed: oidc: expected audience \"#{ENV['JWT_AUDIENCE_ID']}\" got [\"#{ENV['APPID_TENANT']}\"]"
    end

  end

  context 'GET /tenants/{tenant_id}/batches' do

    it 'Success Without Status' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, 'size=1000', {'Authorization' => "Bearer #{@token_all_roles}"})
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
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "status=#{STATUS}&size=1000", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(batch['status']).to eql STATUS
      end
    end

    it 'Success With Name' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "name=#{@batch_name}&size=1000", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(batch['name']).to eql @batch_name
      end
    end

    it 'Success With Greater Than Date' do
      greater_than_date = Date.today - 365
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "gteDate=#{greater_than_date}&size=1000", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(DateTime.strptime(batch['startDate'], '%Y-%m-%dT%H:%M:%S%Z')).to be > greater_than_date
      end
    end

    it 'Success With Less Than Date' do
      less_than_date = Date.today + 1
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "lteDate=#{less_than_date}&size=1000", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['total']).to be > 0
      parsed_response['results'].each do |batch|
        expect(DateTime.strptime(batch['startDate'], '%Y-%m-%dT%H:%M:%S%Z')).to be < less_than_date
      end
    end

    it 'Name Not Found' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "name=#{INVALID_ID}", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be true
    end

    it 'Status Not Found' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "status=#{INVALID_ID}", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be true
    end

    it 'Greater Than Date With No Results' do
      greater_than_date = Date.today + 10000
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "gteDate=#{greater_than_date}", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be true
    end

    it 'Less Than Date With No Results' do
      less_than_date = Date.today - 5000
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "lteDate=#{less_than_date}", {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'].empty?).to be true
    end

    it 'Invalid Greater Than Date' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "gteDate=#{INVALID_ID}", {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Get batch failed: [400] search_phase_execution_exception: all shards failed: parse_exception: failed to parse date field [#{INVALID_ID}] with format [strict_date_optional_time||epoch_millis]: [failed to parse date field [#{INVALID_ID}] with format [strict_date_optional_time||epoch_millis]]"
    end

    it 'Invalid Less Than Date' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, "lteDate=#{INVALID_ID}", {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Get batch failed: [400] search_phase_execution_exception: all shards failed: parse_exception: failed to parse date field [#{INVALID_ID}] with format [strict_date_optional_time||epoch_millis]: [failed to parse date field [#{INVALID_ID}] with format [strict_date_optional_time||epoch_millis]]"
    end

    it 'Query Parameter With Restricted Characters' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, 'status="[{started}]"', {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- status (request query parameter) must not contain the following characters: \"=<>[]{}"
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_get_batches(nil, 'size=1000', {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Integrator ID can not view batches created with a different Integrator ID' do
      #Create Batch
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        dataType: DATA_TYPE,
        topic: BATCH_INPUT_TOPIC,
        metadata: {
          rspec1: 'test1',
          rspec2: 'test2',
          rspec3: {
            rspec3A: 'test3A',
            rspec3B: 'test3B'
          }
        }
      }

      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Modify Batch Integrator ID
      update_batch_script = {
        script: {
          source: 'ctx._source.integratorId = params.integratorId',
          lang: 'painless',
          params: {
            integratorId: 'modified-integrator-id'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @new_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch Integrator ID updated to "modified-integrator-id"')

      #Verify Integrator ID Modified
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @new_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['integratorId']).to eql 'modified-integrator-id'

      #Verify Batch Not Visible to Different Integrator ID
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, 'size=1000', {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      unless parsed_response['results'].empty?
        parsed_response['results'].each do |batch|
          raise "Batch ID #{@new_batch_id} found with different Integrator ID!" if batch['id'] == @new_batch_id
        end
      end

      #Verify Batch Visible To Consumer Role
      @batch_found = false
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, 'size=1000', {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      parsed_response['results'].each do |batch|
        if batch['id'] == @new_batch_id
          @batch_found = true
          expect(batch['integratorId']).to eql 'modified-integrator-id'
        end
      end
      expect(@batch_found).to be true
    end

    it 'Unauthorized - Missing Authorization' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, nil, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{EXISTING_TENANT_ID}' is not included in the authorized scopes: ."
    end

    it 'Unauthorized - No Roles' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, nil, {'Authorization' => "Bearer #{@token_no_roles}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'The access token must have one of these scopes: hri_consumer, hri_data_integrator'
    end

    it 'Unauthorized - Invalid Audience' do
      response = @mgmt_api_helper.hri_get_batches(EXISTING_TENANT_ID, nil, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Authorization token validation failed: oidc: expected audience \"#{ENV['JWT_AUDIENCE_ID']}\" got [\"#{ENV['APPID_TENANT']}\"]"
    end

  end

  context 'GET /tenants/{tenantId}/batches/{batchId}' do

    it 'Success With Consumer Role Only' do
      response = @mgmt_api_helper.hri_get_batch(TENANT_ID, @batch_id, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['id']).to eq @batch_id
      expect(parsed_response['name']).to eql @batch_name
      expect(parsed_response['status']).to eql STATUS
      expect(parsed_response['startDate']).to eql @start_date.to_s
      expect(parsed_response['dataType']).to eql DATA_TYPE
      expect(parsed_response['topic']).to eql BATCH_INPUT_TOPIC
      expect(parsed_response['recordCount']).to eql 1
      expect(parsed_response['metadata']['rspec1']).to eql 'test1'
      expect(parsed_response['metadata']['rspec2']).to eql 'test2'
      expect(parsed_response['metadata']['rspec3']['rspec3A']).to eql 'test3A'
      expect(parsed_response['metadata']['rspec3']['rspec3B']).to eql 'test3B'
    end

    it 'Success With Integrator Role Only' do
      #Create Batch
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        dataType: DATA_TYPE,
        topic: BATCH_INPUT_TOPIC,
        metadata: {
          rspec1: 'test1',
          rspec2: 'test2',
          rspec3: {
            rspec3A: 'test3A',
            rspec3B: 'test3B'
          }
        }
      }
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Get Batch
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @new_batch_id, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['id']).to eq @new_batch_id
      expect(parsed_response['name']).to eql @batch_name
      expect(parsed_response['status']).to eql STATUS
      expect(parsed_response['dataType']).to eql DATA_TYPE
      expect(parsed_response['topic']).to eql BATCH_INPUT_TOPIC
      expect(parsed_response['metadata']['rspec1']).to eql 'test1'
      expect(parsed_response['metadata']['rspec2']).to eql 'test2'
      expect(parsed_response['metadata']['rspec3']['rspec3A']).to eql 'test3A'
      expect(parsed_response['metadata']['rspec3']['rspec3B']).to eql 'test3B'
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_get_batch(nil, @batch_id, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Batch ID Not Found' do
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, INVALID_ID, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The document for tenantId: #{EXISTING_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Integrator ID can not view a batch created with a different Integrator ID' do
      #Create Batch
      @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @batch_name,
        dataType: DATA_TYPE,
        topic: BATCH_INPUT_TOPIC,
        metadata: {
          rspec1: 'test1',
          rspec2: 'test2',
          rspec3: {
            rspec3A: 'test3A',
            rspec3B: 'test3B'
          }
        }
      }
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @new_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@new_batch_id}")

      #Modify Batch Integrator ID
      update_batch_script = {
        script: {
          source: 'ctx._source.integratorId = params.integratorId',
          lang: 'painless',
          params: {
            integratorId: 'modified-integrator-id'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @new_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch Integrator ID updated to "modified-integrator-id"')

      #Verify Integrator ID Modified
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @new_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['integratorId']).to eql 'modified-integrator-id'

      #Verify Batch Not Visible to Different Integrator ID
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @new_batch_id, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include 'does not match the data integratorId'
    end

    it 'Unauthorized - Missing Authorization' do
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @batch_id)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @batch_id, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{EXISTING_TENANT_ID}' is not included in the authorized scopes: ."
    end

    it 'Unauthorized - No Roles' do
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @batch_id, {'Authorization' => "Bearer #{@token_no_roles}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'The access token must have one of these scopes: hri_consumer, hri_data_integrator'
    end

    it 'Unauthorized - Invalid Audience' do
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @batch_id, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Authorization token validation failed: oidc: expected audience \"#{ENV['JWT_AUDIENCE_ID']}\" got [\"#{ENV['APPID_TENANT']}\"]"
    end

    it 'Missing Batch ID With Ending Forward Slash' do
      response = @mgmt_api_helper.hri_custom_request("/tenants/#{TEST_TENANT_ID}/batches//", nil, {'Authorization' => "Bearer #{@token_consumer_role_only}"}, 'GET')
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
          rspec1: 'test3',
          rspec2: 'test4',
          rspec4: {
            rspec4A: 'test4A',
            rspec4B: 'テスト'
          }
        }
      }
      @batch_template = {
        name: @batch_name,
        dataType: DATA_TYPE,
        topic: BATCH_INPUT_TOPIC,
        metadata: {
          rspec1: 'test1',
          rspec2: 'test2',
          rspec3: {
            rspec3A: 'test3A',
            rspec3B: 'test3B'
          }
        }
      }
    end

    it 'Success' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Set Batch Complete
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Complete
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @send_complete_batch_id, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      expect(parsed_response['endDate']).to_not be_nil

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@send_complete_batch_id} and status: completed")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @send_complete_batch_id && parsed_message['status'] == 'completed'
            @message_found = true
            expect(parsed_message['dataType']).to eql DATA_TYPE
            expect(parsed_message['id']).to eql @send_complete_batch_id
            expect(parsed_message['name']).to eql @batch_name
            expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
            expect(parsed_message['status']).to eql 'completed'
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(parsed_message['metadata']['rspec1']).to eql 'test3'
            expect(parsed_message['metadata']['rspec2']).to eql 'test4'
            expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql 'test4A'
            expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql 'テスト'
            expect(parsed_message['metadata']['rspec3']).to be_nil
            expect(parsed_message['expectedRecordCount']).to eq 1
            expect(parsed_message['recordCount']).to eq 1
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Success with recordCount' do
      record_count = {
        recordCount: 1,
        metadata: {
          rspec1: 'test3',
          rspec2: 'test4',
          rspec4: {
            rspec4A: 'test4A',
            rspec4B: 'テスト'
          }
        }
      }

      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Set Batch Complete
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @send_complete_batch_id, 'sendComplete', record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Complete
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @send_complete_batch_id, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      expect(parsed_response['endDate']).to_not be_nil

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@send_complete_batch_id} and status: completed")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @send_complete_batch_id && parsed_message['status'] == 'completed'
            @message_found = true
            expect(parsed_message['dataType']).to eql DATA_TYPE
            expect(parsed_message['id']).to eql @send_complete_batch_id
            expect(parsed_message['name']).to eql @batch_name
            expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
            expect(parsed_message['status']).to eql 'completed'
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(parsed_message['metadata']['rspec1']).to eql 'test3'
            expect(parsed_message['metadata']['rspec2']).to eql 'test4'
            expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql 'test4A'
            expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql 'テスト'
            expect(parsed_message['metadata']['rspec3']).to be_nil
            expect(parsed_message['expectedRecordCount']).to eq 1
            expect(parsed_message['recordCount']).to eq 1
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Invalid Batch ID' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, INVALID_ID, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{EXISTING_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing Record Count' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', nil, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- expectedRecordCount (json field in request body) must be present if recordCount (json field in request body) is not present\n- recordCount (json field in request body) must be present if expectedRecordCount (json field in request body) is not present"
    end

    it 'Invalid Record Count' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', {expectedRecordCount: "1"}, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"expectedRecordCount\": expected type int, but received type string"
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_put_batch(nil, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, nil, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Missing Batch ID and Record Count' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, nil, 'sendComplete', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- expectedRecordCount (json field in request body) must be present if recordCount (json field in request body) is not present\n- id (url path parameter) is a required field\n- recordCount (json field in request body) must be present if expectedRecordCount (json field in request body) is not present"
    end

    it 'Conflict: Batch with a status other than started' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Terminated Status
      update_batch_script = {
        script: {
          source: 'ctx._source.status = params.status',
          lang: 'painless',
          params: {
            status: 'terminated'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql 'terminated'

      #Attempt to complete batch
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'terminated' state"

      #Delete batch
      response = @elastic.es_delete_batch(EXISTING_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
    end

    it 'Conflict: Batch that already has a completed status' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Completed Status
      update_batch_script = {
        script: {
          source: 'ctx._source.status = params.status',
          lang: 'painless',
          params: {
            status: 'completed'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql 'completed'

      #Attempt to complete batch
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'completed' state"
    end

    it 'Integrator ID can not update batches created with a different Integrator ID' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Modify Batch Integrator ID
      update_batch_script = {
        script: {
          source: 'ctx._source.integratorId = params.integratorId',
          lang: 'painless',
          params: {
            integratorId: 'modified-integrator-id'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch Integrator ID updated to "modified-integrator-id"')

      #Verify Integrator ID Modified
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['integratorId']).to eql 'modified-integrator-id'

      #Verify Batch Not Updated With Different Integrator ID
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "but owned by 'modified-integrator-id'"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{EXISTING_TENANT_ID}' is not included in the authorized scopes: ."
    end

    it 'Unauthorized - No Roles' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_no_roles}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to initiate sendComplete on a batch'
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to initiate sendComplete on a batch'
    end

    it 'Unauthorized - Invalid Audience' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Authorization token validation failed: oidc: expected audience \"#{ENV['JWT_AUDIENCE_ID']}\" got [\"#{ENV['APPID_TENANT']}\"]"
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/terminate' do

    before(:all) do
      @terminate_batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
      @batch_template = {
        name: @terminate_batch_name,
        dataType: DATA_TYPE,
        topic: BATCH_INPUT_TOPIC,
        metadata: {
          rspec1: 'test1',
          rspec2: 'test2',
          rspec3: {
            rspec3A: 'test3A',
            rspec3B: 'test3B'
          }
        }
      }
      @terminate_metadata = {
        metadata: {
          rspec1: 'test3',
          rspec2: 'test4',
          rspec4: {
            rspec4A: 'test4A',
            rspec4B: 'テスト'
          }
        }
      }
    end

    it 'Success' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Terminate Batch
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Terminated
      response = @mgmt_api_helper.hri_get_batch(EXISTING_TENANT_ID, @terminate_batch_id, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      expect(parsed_response['endDate']).to_not be_nil

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@terminate_batch_id} and status: terminated")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @terminate_batch_id && parsed_message['status'] == 'terminated'
            @message_found = true
            expect(parsed_message['dataType']).to eql DATA_TYPE
            expect(parsed_message['id']).to eql @terminate_batch_id
            expect(parsed_message['name']).to eql @terminate_batch_name
            expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
            expect(parsed_message['status']).to eql 'terminated'
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
            expect(parsed_message['metadata']['rspec1']).to eql 'test3'
            expect(parsed_message['metadata']['rspec2']).to eql 'test4'
            expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql 'test4A'
            expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql 'テスト'
            expect(parsed_message['metadata']['rspec3']).to be_nil
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Invalid Batch ID' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, INVALID_ID, 'terminate', nil, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{EXISTING_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing Tenant ID' do
      response = @mgmt_api_helper.hri_put_batch(nil, @batch_id, 'terminate', @expected_record_count, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, nil, 'terminate', @expected_record_count, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Conflict: Batch with a status other than started' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Completed Status
      update_batch_script = {
        script: {
          source: 'ctx._source.status = params.status',
          lang: 'painless',
          params: {
            status: 'completed'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql 'completed'

      #Attempt to terminate batch
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'completed' state"

      #Delete batch
      response = @elastic.es_delete_batch(EXISTING_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
    end

    it 'Conflict: Batch that already has a terminated status' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Completed Status
      update_batch_script = {
        script: {
          source: 'ctx._source.status = params.status',
          lang: 'painless',
          params: {
            status: 'terminated'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql 'terminated'

      #Attempt to terminate batch
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'terminated' state"
    end

    it 'Integrator ID can not update batches created with a different Integrator ID' do
      #Create Batch
      response = @mgmt_api_helper.hri_post_batch(EXISTING_TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Modify Batch Integrator ID
      update_batch_script = {
        script: {
          source: 'ctx._source.integratorId = params.integratorId',
          lang: 'painless',
          params: {
            integratorId: 'modified-integrator-id'
          }
        }
      }.to_json
      response = @elastic.es_batch_update(EXISTING_TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql 'updated'
      Logger.new(STDOUT).info('Batch Integrator ID updated to "modified-integrator-id"')

      #Verify Integrator ID Modified
      response = @elastic.es_get_batch(EXISTING_TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['integratorId']).to eql 'modified-integrator-id'

      #Verify Batch Not Updated With Different Integrator ID
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include "but owned by 'modified-integrator-id'"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'terminate', nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Authorization token validation failed: oidc: malformed jwt: square/go-jose: compact JWS format must have three parts'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{EXISTING_TENANT_ID}' is not included in the authorized scopes: ."
    end

    it 'Unauthorized - No Roles' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_no_roles}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to terminate a batch'
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Must have hri_data_integrator role to terminate a batch'
    end

    it 'Unauthorized - Invalid Audience' do
      response = @mgmt_api_helper.hri_put_batch(EXISTING_TENANT_ID, @batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Authorization token validation failed: oidc: expected audience \"#{ENV['JWT_AUDIENCE_ID']}\" got [\"#{ENV['APPID_TENANT']}\"]"
    end

  end

end