# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

describe 'HRI Management API With Validation' do

  INVALID_ID = 'INVALID'
  TENANT_ID = 'test0211'
  AUTHORIZED_TENANT_ID = 'provider1234'
  TENANT_ID_WITH_NO_ROLES = 'provider237'
  INTEGRATOR_ID = 'claims'
  TEST_TENANT_ID = "rspec-#{'-'.delete('.')}-test-tenant".downcase
  TEST_INTEGRATOR_ID = "rspec-#{'-'.delete('.')}-test-integrator".downcase
  DATA_TYPE = 'rspec-batch'
  STATUS = 'started'
  COMPLETED_BATCH_ID = '9Swf82h1LBygCb0tbklQ'
  BATCH_INPUT_TOPIC = "ingest.#{AUTHORIZED_TENANT_ID}.#{INTEGRATOR_ID}.in"
  KAFKA_TIMEOUT = 60
  INVALIDTHRESHOLD = 5
  INVALID_RECORD_COUNT = 3
  ACTUAL_RECORD_COUNT = 15
  EXPECTED_RECORD_COUNT = 15
  FAILURE_MESSAGE = 'Rspec Failure Message'


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

  def hri_get_batch(tenant_id, batch_id, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches/#{batch_id}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_get_batches(tenant_id, query_params = nil, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches"
    url += "?#{query_params}" unless query_params.nil?
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}" }.merge(override_headers)
    rest_get(url, headers)
  end

  def hri_put_batch(tenant_id, batch_id, action, additional_params = {}, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches/#{batch_id}/action/#{action}"
    @az_token = get_access_token()
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@az_token}"}.merge(override_headers)
    rest_put(url, additional_params.to_json, headers)
  end

  def hri_post_batch(tenant_id, request_body, override_headers = {})
    url = "#{@hri_base_url}/tenants/#{tenant_id}/batches"
    @az_token = get_access_token()
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
    @iam_token = get_access_token()
    # @elastic = HRITestHelpers::ElasticHelper.new({url: ENV['ELASTIC_URL'], username: ENV['ELASTIC_USERNAME'], password: ENV['ELASTIC_PASSWORD']})
    # @iam_token = HRITestHelpers::IAMHelper.new(ENV['IAM_CLOUD_URL']).get_access_token(ENV['CLOUD_API_KEY'])
    # @mgmt_api_helper = HRITestHelpers::MgmtAPIHelper.new(@hri_base_url, @iam_token)
    # @hri_deploy_helper = HRIDeployHelper.new
    # @event_streams_api_helper = HRITestHelpers::EventStreamsAPIHelper.new(ENV['ES_ADMIN_URL'], ENV['CLOUD_API_KEY'])
    # @app_id_helper = HRITestHelpers::AppIDHelper.new(ENV['APPID_URL'], ENV['APPID_TENANT'], @iam_token, ENV['JWT_AUDIENCE_ID'])
    @start_date = DateTime.now

    @exe_path = File.absolute_path(File.join(File.dirname(__FILE__), "../../src/hri"))
    @config_path = File.absolute_path(File.join(File.dirname(__FILE__), "test_config"))
    @log_path = File.absolute_path(File.join(File.dirname(__FILE__), "../logs"))
    Dir.mkdir(@log_path) unless Dir.exists?(@log_path)

    # @hri_deploy_helper.deploy_hri(@exe_path, "#{@config_path}/valid_config.yml", @log_path, 'validation-', '-validation=true')
    # response = @request_helper.rest_get("#{@hri_base_url}/healthcheck", {})
    # unless response.code == 200
    #   raise "Health check failed: #{response.body}"
    # end

    #Initialize Kafka Consumer
    # @kafka = Kafka.new(ENV['KAFKA_BROKERS'], sasl_plain_username: 'token', sasl_plain_password: ENV['KAFKA_PASSWORD'], ssl_ca_certs_from_system: true)
    # @kafka_consumer = @kafka.consumer(group_id: 'rspec-mgmt-api-consumer')
    # @kafka_consumer.subscribe("ingest.#{TENANT_ID}.#{INTEGRATOR_ID}.notification")

    #Create Batch
    @batch_prefix = "rspec-batch"
    @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
    create_batch = {
      name: @batch_name,
      topic: BATCH_INPUT_TOPIC,
      #status: STATUS,
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

    #Get AppId Access Tokens
    # @token_invalid_tenant = @app_id_helper.get_access_token('hri_integration_tenant_test_invalid', 'tenant_test_invalid')
    # @token_no_roles = @app_id_helper.get_access_token('hri_integration_tenant_test', 'tenant_test')
    # @token_integrator_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_integrator', 'tenant_test hri_data_integrator')
    # @token_consumer_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_consumer', 'tenant_test hri_consumer')
    # @token_all_roles = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer')
    # @token_internal_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_internal', 'tenant_test hri_internal')
    # @token_invalid_audience = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer', ENV['APPID_TENANT'])
  end


  context 'POST /tenants/{tenant_id}/streams/{integrator_id}' do

    before(:all) do
      @stream_info = {
        numPartitions: 1,
        retentionMs: 3600000
      }
    end

    it 'Success' do
      #Create Tenant
      response = hri_post_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['tenantId']).to eql TEST_TENANT_ID

      #Create Stream
      response = hri_post_tenant_stream(TEST_TENANT_ID, TEST_INTEGRATOR_ID, @stream_info.to_json)
      expect(response.code).to eq 201

      #Verify Stream Creation
      response = hri_get_tenant_streams(TEST_TENANT_ID)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['results'][0]['id']).to eql TEST_INTEGRATOR_ID

      # Timeout.timeout(30, nil, 'Kafka topics not created after 30 seconds') do
      #   loop do
      #     topics = @event_streams_api_helper.get_topics
      #     break if (topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.in") &&
      #               topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.notification") &&
      #               topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.out") &&
      #               topics.include?("ingest.#{TEST_TENANT_ID}.#{TEST_INTEGRATOR_ID}.invalid"))
      #   end
      # end
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

      #Delete Tenant
      response = hri_delete_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 200

      #Verify Tenant Deleted
      response = hri_get_tenant(TEST_TENANT_ID)
      expect(response.code).to eq 404
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/sendComplete' do

    before(:all) do
      @expected_record_count = {
        expectedRecordCount: EXPECTED_RECORD_COUNT,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 15
        }
      }

      @terminate_metadata = {
        "metadata": {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      @record_counts_and_message = {
        actualRecordCount: ACTUAL_RECORD_COUNT,
        failureMessage: FAILURE_MESSAGE,
        invalidRecordCount: INVALID_RECORD_COUNT
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

      #Set Batch to Send Completed
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 200

      #Verify Batch Send Completed
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'sendCompleted'
      expect(parsed_response['endDate']).to be_nil
      expect(parsed_response['expectedRecordCount']).to eq EXPECTED_RECORD_COUNT
      expect(parsed_response['recordCount']).to eq EXPECTED_RECORD_COUNT

      # #Verify Kafka Message
      # Timeout.timeout(KAFKA_TIMEOUT) do
      #   Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@send_complete_batch_id} and status: sendCompleted")
      #   @kafka_consumer.each_message do |message|
      #     parsed_message = JSON.parse(message.value)
      #     if parsed_message['id'] == @send_complete_batch_id && parsed_message['status'] == 'sendCompleted'
      #       @message_found = true
      #       expect(parsed_message['dataType']).to eql DATA_TYPE
      #       expect(parsed_message['id']).to eql @send_complete_batch_id
      #       expect(parsed_message['name']).to eql @batch_name
      #       expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
      #       expect(parsed_message['invalidThreshold']).to eql INVALID_THRESHOLD
      #       expect(parsed_message['expectedRecordCount']).to eq EXPECTED_RECORD_COUNT
      #       expect(parsed_message['recordCount']).to eq EXPECTED_RECORD_COUNT
      #       expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(parsed_message['metadata']['rspec1']).to eql 'test3'
      #       expect(parsed_message['metadata']['rspec2']).to eql 'test4'
      #       expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql 'test4A'
      #       expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql 'テスト'
      #       expect(parsed_message['metadata']['rspec3']).to be_nil
      #       break
      #     end
      #   end
      #   expect(@message_found).to be true
      # end
    end

    it 'Invalid Tenant ID' do
      response = hri_put_batch(INVALID_ID, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{INVALID_ID}' is not included in the authorized roles:tenant_#{INVALID_ID}."
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

    it 'Conflict: Batch with a status of completed' do
      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, COMPLETED_BATCH_ID)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      Logger.new(STDOUT).info("Batch Id '#{COMPLETED_BATCH_ID}' is in completed state")

      #Attempt to send complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, COMPLETED_BATCH_ID, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'completed' state"
    end

    it 'Conflict: Batch with a status of terminated' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@send_complete_batch_id}")

      #Changing batch status to "terminated"
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id,'terminate', @terminate_metadata)
      expect(change_status.code).to eq 200
      Logger.new(STDOUT).info("Batch status updated to terminated")

      #Verify Updated Batch Status
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'

      #Attempt to send complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'terminated' state"
     end

    it 'Conflict: Batch with a status of failed' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Failed Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id,'fail', @record_counts_and_message)
      expect(change_status.code).to eq 200
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'failed'

      #Attempt to send complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'failed' state"
    end

    it 'Conflict: Batch that already has a sendCompleted status' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Completed Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id,'sendComplete', @expected_record_count)
      expect(change_status.code).to eq 200
      parsed_response = JSON.parse(response.body)
      Logger.new(STDOUT).info('Batch status updated to "sendCompleted"')

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'sendCompleted'

      #Attempt to send complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "sendComplete failed, batch is in 'sendCompleted' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count,{'Authorization' => nil })
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
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Azure AD authentication returned 401"
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/processingComplete' do

    before(:all) do
      @record_counts = {
        invalidRecordCount: INVALID_RECORD_COUNT,
        actualRecordCount: ACTUAL_RECORD_COUNT
      }

      @expected_record_count = {
        expectedRecordCount: EXPECTED_RECORD_COUNT,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 15
        }
      }

      @terminate_metadata = {
        "metadata": {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      @record_counts_and_message = {
        actualRecordCount: ACTUAL_RECORD_COUNT,
        failureMessage: FAILURE_MESSAGE,
        invalidRecordCount: INVALID_RECORD_COUNT
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
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Update Batch to sendCompleted Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id,'sendComplete', @expected_record_count)
      expect(change_status.code).to eq 200
      Logger.new(STDOUT).info('Batch status updated to "sendCompleted"')

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'sendCompleted'

      #Set Batch to Completed
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 200
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Processing Completed
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      expect(parsed_response['endDate']).to_not be_nil
      expect(parsed_response['invalidRecordCount']).to eq INVALID_RECORD_COUNT

      #Verify Kafka Message
      # Timeout.timeout(KAFKA_TIMEOUT) do
      #   Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@processing_complete_batch_id} and status: completed")
      #   @kafka_consumer.each_message do |message|
      #     parsed_message = JSON.parse(message.value)
      #     if parsed_message['id'] == @processing_complete_batch_id && parsed_message['status'] == 'completed'
      #       @message_found = true
      #       expect(parsed_message['dataType']).to eql DATA_TYPE
      #       expect(parsed_message['id']).to eql @processing_complete_batch_id
      #       expect(parsed_message['name']).to eql @batch_name
      #       expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
      #       expect(parsed_message['invalidThreshold']).to eql INVALID_THRESHOLD
      #       expect(parsed_message['invalidRecordCount']).to eql INVALID_RECORD_COUNT
      #       expect(parsed_message['actualRecordCount']).to eql ACTUAL_RECORD_COUNT
      #       expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(parsed_message['metadata']['rspec1']).to eql 'test1'
      #       expect(parsed_message['metadata']['rspec2']).to eql 'test2'
      #       expect(parsed_message['metadata']['rspec3']['rspec3A']).to eql 'test3A'
      #       expect(parsed_message['metadata']['rspec3']['rspec3B']).to eql 'test3B'
      #       break
      #     end
      #   end
      #   expect(@message_found).to be true
      # end
    end

    it 'Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Invalid Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, INVALID_ID, 'processingComplete', @record_counts)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{AUTHORIZED_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing invalidRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'processingComplete', {actualRecordCount: ACTUAL_RECORD_COUNT})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- invalidRecordCount (json field in request body) is a required field"
    end

    it 'Invalid invalidRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'processingComplete', {invalidRecordCount: "1", actualRecordCount: ACTUAL_RECORD_COUNT})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"invalidRecordCount\": expected type int, but received type string"
    end

    it 'Missing actualRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'processingComplete', {invalidRecordCount: INVALID_RECORD_COUNT})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- actualRecordCount (json field in request body) is a required field"
    end

    it 'Invalid actualRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'processingComplete', {actualRecordCount: "1", invalidRecordCount: INVALID_RECORD_COUNT})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"actualRecordCount\": expected type int, but received type string"
    end

    it 'Missing Tenant ID' do
      response = hri_put_batch(nil, @batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'processingComplete', @record_counts)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Missing Batch ID and actualRecordCount' do
      @record_counts.delete(:actualRecordCount)
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'processingComplete', @record_counts)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- actualRecordCount (json field in request body) is a required field\n- id (url path parameter) is a required field"
      @record_counts[:actualRecordCount] = ACTUAL_RECORD_COUNT
    end

    it 'Conflict: Batch with a status of started' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Attempt to process complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "processingComplete failed, batch is in 'started' state"
    end

    it 'Conflict: Batch with a status of terminated' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Update Batch to Terminated Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id,'terminate', @terminate_metadata)
      expect(change_status.code).to eq 200

      #Verify Updated Batch Status
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Attempt to processing complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "processingComplete failed, batch is in 'terminated' state"
    end

    it 'Conflict: Batch with a status of failed' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Update Batch to Failed Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id,'fail', @record_counts_and_message)
      expect(change_status.code).to eq 200
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'failed'

      #Attempt to process complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "processingComplete failed, batch is in 'failed' state"
    end

    it 'Conflict: Batch that already has a completed status' do
       #Create Batch
      # response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      # expect(response.code).to eq 201
      # parsed_response = JSON.parse(response.body)
      # @processing_complete_batch_id = parsed_response['id']
      # Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, COMPLETED_BATCH_ID)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      Logger.new(STDOUT).info("Batch Id '#{COMPLETED_BATCH_ID}' is in completed state")

      #Attempt to process complete batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, COMPLETED_BATCH_ID, 'processingComplete', @record_counts)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "processingComplete failed, batch is in 'completed' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'processingComplete', @record_counts,{'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'processingComplete', @record_counts,{'Authorization' => "INVALID"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/terminate' do

    before(:all) do
      @record_counts = {
        invalidRecordCount: INVALID_RECORD_COUNT,
        actualRecordCount: ACTUAL_RECORD_COUNT
      }

      @expected_record_count = {
        expectedRecordCount: EXPECTED_RECORD_COUNT,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 15
        }
      }

      @terminate_metadata = {
        "metadata": {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      @record_counts_and_message = {
        actualRecordCount: ACTUAL_RECORD_COUNT,
        failureMessage: FAILURE_MESSAGE,
        invalidRecordCount: INVALID_RECORD_COUNT
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
      Logger.new(STDOUT).info("Batch Id : #{@terminate_batch_id} is in 'terminated' state")

      #Verify Kafka Message
      # Timeout.timeout(KAFKA_TIMEOUT) do
      #   Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@terminate_batch_id} and status: terminated")
      #   @kafka_consumer.each_message do |message|
      #     parsed_message = JSON.parse(message.value)
      #     if parsed_message['id'] == @terminate_batch_id && parsed_message['status'] == 'terminated'
      #       @message_found = true
      #       expect(parsed_message['dataType']).to eql DATA_TYPE
      #       expect(parsed_message['id']).to eql @terminate_batch_id
      #       expect(parsed_message['name']).to eql @batch_name
      #       expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
      #       expect(parsed_message['invalidThreshold']).to eql INVALID_THRESHOLD
      #       expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(parsed_message['metadata']['rspec1']).to eql 'test3'
      #       expect(parsed_message['metadata']['rspec2']).to eql 'test4'
      #       expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql 'test4A'
      #       expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql 'テスト'
      #       expect(parsed_message['metadata']['rspec3']).to be_nil
      #       break
      #     end
      #   end
      #   expect(@message_found).to be true
      # end
    end

    it 'Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'terminate', nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Invalid Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, INVALID_ID, 'terminate', nil)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{AUTHORIZED_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing Tenant ID' do
      response = hri_put_batch(nil, @batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'terminate', @terminate_metadata)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Conflict: Batch with a status of sendCompleted' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to sendCompleted Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id,'sendComplete', @expected_record_count)
      expect(change_status.code).to eq 200

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'sendCompleted'
      Logger.new(STDOUT).info("Batch ID: #{@terminate_batch_id} is in 'sendCompleted' state")

      #Attempt to terminate batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'sendCompleted' state"
    end

    it 'Conflict: Batch with a status of completed' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, COMPLETED_BATCH_ID)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      Logger.new(STDOUT).info("Batch Id '#{COMPLETED_BATCH_ID}' is in completed state")

      #Attempt to terminate batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, COMPLETED_BATCH_ID, 'terminate', @terminate_metadata)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'completed' state"
    end

    it 'Conflict: Batch with a status of failed' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Failed Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id,'fail', @record_counts_and_message)
      expect(change_status.code).to eq 200

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'failed'
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Attempt to terminate batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'failed' state"
    end

    it 'Conflict: Batch that already has a terminated status' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Terminated Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id,'terminate', @terminate_metadata)
      expect(change_status.code).to eq 200

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Attempt to terminate batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "terminate failed, batch is in 'terminated' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'terminate', @terminate_metadata,{'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - No Roles' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Attempt to terminate batch
      response = hri_put_batch(TENANT_ID_WITH_NO_ROLES, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{TENANT_ID_WITH_NO_ROLES} with document (batch) ID: #{@terminate_batch_id} was not found"
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json )
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Attempt to terminate batch
      response = hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - Invalid Audience' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Attempt to terminate batch
      response = hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/fail' do

    before(:all) do
      @record_counts = {
        invalidRecordCount: INVALID_RECORD_COUNT,
        actualRecordCount: ACTUAL_RECORD_COUNT
      }

      @expected_record_count = {
        expectedRecordCount: EXPECTED_RECORD_COUNT,
        metadata: {
          "compression": "gzip",
          "finalRecordCount": 15
        }
      }

      @terminate_metadata = {
        "metadata": {
          "compression": "gzip",
          "finalRecordCount": 20
        }
      }

      @record_counts_and_message = {
        actualRecordCount: ACTUAL_RECORD_COUNT,
        failureMessage: FAILURE_MESSAGE,
        invalidRecordCount: INVALID_RECORD_COUNT
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
      @failed_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Failed Batch Created With ID: #{@failed_batch_id}")

      #Set Batch to Failed
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 200
      Logger.new(STDOUT).info("Batch ID: #{@failed_batch_id} updated to 'failed' state")

      #Verify Batch Failed
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @failed_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'failed'
      expect(parsed_response['endDate']).to_not be_nil
      expect(parsed_response['invalidRecordCount']).to eq INVALID_RECORD_COUNT

      #Verify Kafka Message
      # Timeout.timeout(KAFKA_TIMEOUT) do
      #   Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@failed_batch_id} and status: failed")
      #   @kafka_consumer.each_message do |message|
      #     parsed_message = JSON.parse(message.value)
      #     if parsed_message['id'] == @failed_batch_id && parsed_message['status'] == 'failed'
      #       @message_found = true
      #       expect(parsed_message['dataType']).to eql DATA_TYPE
      #       expect(parsed_message['id']).to eql @failed_batch_id
      #       expect(parsed_message['name']).to eql @batch_name
      #       expect(parsed_message['topic']).to eql BATCH_INPUT_TOPIC
      #       expect(parsed_message['invalidThreshold']).to eql INVALID_THRESHOLD
      #       expect(parsed_message['invalidRecordCount']).to eql INVALID_RECORD_COUNT
      #       expect(parsed_message['actualRecordCount']).to eql ACTUAL_RECORD_COUNT
      #       expect(parsed_message['failureMessage']).to eql FAILURE_MESSAGE
      #       expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq Date.today.strftime("%Y-%m-%d")
      #       expect(parsed_message['metadata']['rspec1']).to eql 'test1'
      #       expect(parsed_message['metadata']['rspec2']).to eql 'test2'
      #       expect(parsed_message['metadata']['rspec3']['rspec3A']).to eql 'test3A'
      #       expect(parsed_message['metadata']['rspec3']['rspec3B']).to eql 'テスト'
      #       break
      #     end
      #   end
      #   expect(@message_found).to be true
      # end
    end

    it 'Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Invalid Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, INVALID_ID, 'fail', @record_counts_and_message)
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "error getting current Batch Status: The document for tenantId: #{AUTHORIZED_TENANT_ID} with document (batch) ID: #{INVALID_ID} was not found"
    end

    it 'Missing invalidRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', {actualRecordCount: ACTUAL_RECORD_COUNT, failureMessage: 'RSpec failure message'})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- invalidRecordCount (json field in request body) is a required field"
    end

    it 'Invalid invalidRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', {invalidRecordCount: "1", actualRecordCount: ACTUAL_RECORD_COUNT, failureMessage: 'RSpec failure message'})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"invalidRecordCount\": expected type int, but received type string"
    end

    it 'Missing actualRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', {invalidRecordCount: INVALID_RECORD_COUNT, failureMessage: 'RSpec failure message'})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- actualRecordCount (json field in request body) is a required field"
    end

    it 'Invalid actualRecordCount' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', {actualRecordCount: "1", invalidRecordCount: INVALID_RECORD_COUNT, failureMessage: 'RSpec failure message'})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"actualRecordCount\": expected type int, but received type string"
    end

    it 'Missing failureMessage' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', {actualRecordCount: ACTUAL_RECORD_COUNT, invalidRecordCount: INVALID_RECORD_COUNT})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- failureMessage (json field in request body) is a required field"
    end

    it 'Invalid failureMessage' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', {invalidRecordCount: INVALID_RECORD_COUNT, actualRecordCount: ACTUAL_RECORD_COUNT, failureMessage: 10})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request param \"failureMessage\": expected type string, but received type number"
    end

    it 'Missing Tenant ID' do
      response = hri_put_batch(nil, @batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- tenantId (url path parameter) is a required field"
    end

    it 'Missing Batch ID' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'fail', @record_counts_and_message)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- id (url path parameter) is a required field"
    end

    it 'Missing Batch ID and failureMessage' do
      @record_counts_and_message.delete(:failureMessage)
      response = hri_put_batch(AUTHORIZED_TENANT_ID, nil, 'fail', @record_counts_and_message)
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "invalid request arguments:\n- failureMessage (json field in request body) is a required field\n- id (url path parameter) is a required field"
      @record_counts_and_message[:failureMessage] = FAILURE_MESSAGE
    end

    it 'Conflict: Batch with a status of terminated' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @failed_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Failed Batch Created With ID: #{@failed_batch_id}")

      #Update Batch to Terminated Status
      change_status = hri_put_batch(AUTHORIZED_TENANT_ID, @failed_batch_id,'terminate', @terminate_metadata)
      expect(change_status.code).to eq 200

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @failed_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'terminated'
      Logger.new(STDOUT).info("Batch ID : #{@failed_batch_id} is in 'terminated' state")

      #Attempt to fail batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "'fail' failed, batch is in 'terminated' state"
    end

    it 'Conflict: Batch that already has a failed status' do
      #Create Batch
      response = hri_post_batch(AUTHORIZED_TENANT_ID, @batch_template.to_json)
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @failed_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Failed Batch Created With ID: #{@failed_batch_id}")

      #Update Batch to Failed Status
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 200
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = hri_get_batch(AUTHORIZED_TENANT_ID, @failed_batch_id)
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'failed'
      Logger.new(STDOUT).info("Batch ID : #{@failed_batch_id} is in 'failed' state")

      #Attempt to fail batch
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "'fail' failed, batch is in 'failed' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', @record_counts_and_message,{'Authorization' => nil })
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Authorization' do
      response = hri_put_batch(AUTHORIZED_TENANT_ID, @batch_id, 'fail', @record_counts_and_message, {'Authorization' => "INVALID"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql 'Azure AD authentication returned 401'
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

    it 'Unauthorized - Invalid Audience' do
      response = hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized roles:tenant_#{TENANT_ID}."
    end

  end

end