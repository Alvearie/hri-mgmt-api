# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require_relative '../env'

describe 'HRI Management API With Validation' do

  INVALID_ID = 'INVALID'
  TENANT_ID = 'test'
  INTEGRATOR_ID = 'claims'
  DATA_TYPE = 'rspec-batch'
  BATCH_INPUT_TOPIC = "ingest.#{TENANT_ID}.#{INTEGRATOR_ID}.in"
  KAFKA_TIMEOUT = 60
  INVALID_THRESHOLD = 5
  INVALID_RECORD_COUNT = 3
  ACTUAL_RECORD_COUNT = 15
  EXPECTED_RECORD_COUNT = 15
  FAILURE_MESSAGE = 'Rspec Failure Message'

  before(:all) do
    @elastic = ElasticHelper.new
    @app_id_helper = AppIDHelper.new
    @hri_helper = HRIHelper.new(`bx fn api list`.scan(/https.*hri/).first)
    @start_date = DateTime.now

    #Initialize Kafka Consumer
    @kafka = Kafka.new(ENV['KAFKA_BROKERS'], sasl_plain_username: 'token', sasl_plain_password: ENV['KAFKA_PASSWORD'], ssl_ca_certs_from_system: true)
    @kafka_consumer = @kafka.consumer(group_id: 'rspec-test-consumer')
    @kafka_consumer.subscribe("ingest.#{TENANT_ID}.#{INTEGRATOR_ID}.notification")

    #Create Batch
    @batch_prefix = "rspec-#{ENV['BRANCH_NAME'].delete('.')}"
    @batch_name = "#{@batch_prefix}-#{SecureRandom.uuid}"
    create_batch = {
        name: @batch_name,
        status: 'started',
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
    response = @elastic.es_create_batch(TENANT_ID, create_batch)
    expect(response.code).to eq 201
    parsed_response = JSON.parse(response.body)
    @batch_id = parsed_response['_id']
    Logger.new(STDOUT).info("New Batch Created With ID: #{@batch_id}")

    #Get AppId Access Tokens
    @token_invalid_tenant = @app_id_helper.get_access_token('hri_integration_tenant_test_invalid', 'tenant_test_invalid')
    @token_no_roles = @app_id_helper.get_access_token('hri_integration_tenant_test', 'tenant_test')
    @token_integrator_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_integrator', 'tenant_test hri_data_integrator')
    @token_consumer_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_data_consumer', 'tenant_test hri_consumer')
    @token_all_roles = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer')
    @token_internal_role_only = @app_id_helper.get_access_token('hri_integration_tenant_test_internal', 'tenant_test hri_internal')
    @token_invalid_audience = @app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer', ENV['APPID_TENANT'])
  end

  after(:all) do
    #Delete Batches
    response = @elastic.es_delete_by_query(TENANT_ID, "name:rspec-#{@batch_prefix}*")
    response.nil? ? (raise 'Elastic batch delete did not return a response') : (expect(response.code).to eq 200)
    Logger.new(STDOUT).info("Delete test batches by query response #{response.body}")
    @kafka_consumer.stop
  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/sendComplete' do

    before(:all) do
      @expected_record_count = {
          expectedRecordCount: EXPECTED_RECORD_COUNT,
          metadata: {
              rspec1: 'test3',
              rspec2: 'test4',
              rspec4: {
                  rspec4A: 'test4A',
                  rspec4B: 'test4B'
              }
          }
      }
      @batch_template = {
          name: @batch_name,
          dataType: DATA_TYPE,
          topic: BATCH_INPUT_TOPIC,
          invalidThreshold: INVALID_THRESHOLD,
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
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Set Batch to Send Completed
      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Send Completed
      response = @hri_helper.hri_get_batch(TENANT_ID, @send_complete_batch_id, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'sendCompleted'
      expect(parsed_response['endDate']).to be_nil
      expect(parsed_response['expectedRecordCount']).to eq EXPECTED_RECORD_COUNT
      expect(parsed_response['recordCount']).to eq EXPECTED_RECORD_COUNT

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@send_complete_batch_id} and status: sendCompleted")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @send_complete_batch_id && parsed_message['status'] == 'sendCompleted'
            @message_found = true
            expect(parsed_message['dataType']).to eql(DATA_TYPE)
            expect(parsed_message['id']).to eql(@send_complete_batch_id)
            expect(parsed_message['name']).to eql(@batch_name)
            expect(parsed_message['topic']).to eql(BATCH_INPUT_TOPIC)
            expect(parsed_message['invalidThreshold']).to eql(INVALID_THRESHOLD)
            expect(parsed_message['expectedRecordCount']).to eq EXPECTED_RECORD_COUNT
            expect(parsed_message['recordCount']).to eq EXPECTED_RECORD_COUNT
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(parsed_message['metadata']['rspec1']).to eql('test3')
            expect(parsed_message['metadata']['rspec2']).to eql('test4')
            expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql('test4A')
            expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql('test4B')
            expect(parsed_message['metadata']['rspec3']).to be_nil
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Invalid Batch ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, INVALID_ID, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include('document_missing_exception')
    end

    it 'Missing Record Count' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', nil, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing required parameter(s): [expectedRecordCount]')
    end

    it 'Invalid Record Count' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', {expectedRecordCount: "1"}, {'Authorization' => "Bearer #{@token_all_roles}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Invalid parameter type(s): [expectedRecordCount must be a float64, got string instead.]')
    end

    it 'Conflict: Batch with a status of completed' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
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
      response = @elastic.es_batch_update(TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('completed')

      #Attempt to send complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'sendComplete' endpoint failed, batch is in 'completed' state"
    end

    it 'Conflict: Batch with a status of terminated' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
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
      response = @elastic.es_batch_update(TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('terminated')

      #Attempt to send complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'sendComplete' endpoint failed, batch is in 'terminated' state"
    end

    it 'Conflict: Batch with a status of failed' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      #Update Batch to Failed Status
      update_batch_script = {
          script: {
              source: 'ctx._source.status = params.status',
              lang: 'painless',
              params: {
                  status: 'failed'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('failed')

      #Attempt to send complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'sendComplete' endpoint failed, batch is in 'failed' state"
    end

    it 'Conflict: Batch that already has a sendCompleted status' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
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
                  status: 'sendCompleted'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @send_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "sendCompleted"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @send_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('sendCompleted')

      #Attempt to send complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'sendComplete' endpoint failed, batch is in 'sendCompleted' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', @expected_record_count)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing Authorization header')
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Unauthorized - No Roles' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_no_roles}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Must have hri_data_integrator role to update a batch')
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Must have hri_data_integrator role to update a batch')
    end

    it 'Unauthorized - Invalid Audience' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @send_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Send Complete Batch Created With ID: #{@send_complete_batch_id}")

      response = @hri_helper.hri_put_batch(TENANT_ID, @send_complete_batch_id, 'sendComplete', @expected_record_count, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/processingComplete' do

    before(:all) do
      @record_counts = {
          invalidRecordCount: INVALID_RECORD_COUNT,
          actualRecordCount: ACTUAL_RECORD_COUNT
      }
      @batch_template = {
          name: @batch_name,
          dataType: DATA_TYPE,
          topic: BATCH_INPUT_TOPIC,
          invalidThreshold: INVALID_THRESHOLD,
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
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Update Batch to sendCompleted Status
      update_batch_script = {
          script: {
              source: 'ctx._source.status = params.status',
              lang: 'painless',
              params: {
                  status: 'sendCompleted'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @processing_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "sendCompleted"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @processing_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('sendCompleted')

      #Set Batch to Completed
      response = @hri_helper.hri_put_batch(TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Processing Completed
      response = @hri_helper.hri_get_batch(TENANT_ID, @processing_complete_batch_id, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'completed'
      expect(parsed_response['endDate']).to_not be_nil
      expect(parsed_response['invalidRecordCount']).to eq INVALID_RECORD_COUNT

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@processing_complete_batch_id} and status: completed")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @processing_complete_batch_id && parsed_message['status'] == 'completed'
            @message_found = true
            expect(parsed_message['dataType']).to eql(DATA_TYPE)
            expect(parsed_message['id']).to eql(@processing_complete_batch_id)
            expect(parsed_message['name']).to eql(@batch_name)
            expect(parsed_message['topic']).to eql(BATCH_INPUT_TOPIC)
            expect(parsed_message['invalidThreshold']).to eql(INVALID_THRESHOLD)
            expect(parsed_message['invalidRecordCount']).to eql(INVALID_RECORD_COUNT)
            expect(parsed_message['actualRecordCount']).to eql(ACTUAL_RECORD_COUNT)
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(parsed_message['metadata']['rspec1']).to eql('test1')
            expect(parsed_message['metadata']['rspec2']).to eql('test2')
            expect(parsed_message['metadata']['rspec3']['rspec3A']).to eql('test3A')
            expect(parsed_message['metadata']['rspec3']['rspec3B']).to eql('test3B')
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Invalid Batch ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, INVALID_ID, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include('document_missing_exception')
    end

    it 'Missing invalidRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', {actualRecordCount: ACTUAL_RECORD_COUNT}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing required parameter(s): [invalidRecordCount]')
    end

    it 'Invalid invalidRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', {invalidRecordCount: "1", actualRecordCount: ACTUAL_RECORD_COUNT}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Invalid parameter type(s): [invalidRecordCount must be a float64, got string instead.]')
    end

    it 'Missing actualRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', {invalidRecordCount: INVALID_RECORD_COUNT}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing required parameter(s): [actualRecordCount]')
    end

    it 'Invalid actualRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', {actualRecordCount: "1", invalidRecordCount: INVALID_RECORD_COUNT}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Invalid parameter type(s): [actualRecordCount must be a float64, got string instead.]')
    end

    it 'Conflict: Batch with a status of started' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Attempt to process complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'processingComplete' endpoint failed, batch is in 'started' state"
    end

    it 'Conflict: Batch with a status of terminated' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

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
      response = @elastic.es_batch_update(TENANT_ID, @processing_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @processing_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('terminated')

      #Attempt to process complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'processingComplete' endpoint failed, batch is in 'terminated' state"
    end

    it 'Conflict: Batch with a status of failed' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

      #Update Batch to Failed Status
      update_batch_script = {
          script: {
              source: 'ctx._source.status = params.status',
              lang: 'painless',
              params: {
                  status: 'failed'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @processing_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @processing_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('failed')

      #Attempt to process complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'processingComplete' endpoint failed, batch is in 'failed' state"
    end

    it 'Conflict: Batch that already has a completed status' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @processing_complete_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Processing Complete Batch Created With ID: #{@processing_complete_batch_id}")

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
      response = @elastic.es_batch_update(TENANT_ID, @processing_complete_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @processing_complete_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('completed')

      #Attempt to process complete batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @processing_complete_batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'processingComplete' endpoint failed, batch is in 'completed' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing Authorization header')
    end

    it 'Unauthorized - Invalid Authorization' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Must have hri_internal role to mark a batch as processing complete')
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Unauthorized - Invalid Audience' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'processingComplete', @record_counts, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/terminate' do

    before(:all) do
      @batch_template = {
          name: @batch_name,
          dataType: DATA_TYPE,
          topic: BATCH_INPUT_TOPIC,
          invalidThreshold: INVALID_THRESHOLD,
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
                  rspec4B: 'test4B'
              }
          }
      }
    end

    it 'Success' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Terminate Batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', @terminate_metadata, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Terminated
      response = @hri_helper.hri_get_batch(TENANT_ID, @terminate_batch_id, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
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
            expect(parsed_message['dataType']).to eql(DATA_TYPE)
            expect(parsed_message['id']).to eql(@terminate_batch_id)
            expect(parsed_message['name']).to eql(@batch_name)
            expect(parsed_message['topic']).to eql(BATCH_INPUT_TOPIC)
            expect(parsed_message['invalidThreshold']).to eql(INVALID_THRESHOLD)
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(parsed_message['metadata']['rspec1']).to eql('test3')
            expect(parsed_message['metadata']['rspec2']).to eql('test4')
            expect(parsed_message['metadata']['rspec4']['rspec4A']).to eql('test4A')
            expect(parsed_message['metadata']['rspec4']['rspec4B']).to eql('test4B')
            expect(parsed_message['metadata']['rspec3']).to be_nil
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Invalid Batch ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, INVALID_ID, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include('document_missing_exception')
    end

    it 'Conflict: Batch with a status of sendCompleted' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to sendCompleted Status
      update_batch_script = {
          script: {
              source: 'ctx._source.status = params.status',
              lang: 'painless',
              params: {
                  status: 'sendCompleted'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "sendCompleted"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('sendCompleted')

      #Attempt to terminate batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'terminate' endpoint failed, batch is in 'sendCompleted' state"
    end

    it 'Conflict: Batch with a status of completed' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
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
      response = @elastic.es_batch_update(TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('completed')

      #Attempt to terminate batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'terminate' endpoint failed, batch is in 'completed' state"
    end

    it 'Conflict: Batch with a status of failed' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      #Update Batch to Failed Status
      update_batch_script = {
          script: {
              source: 'ctx._source.status = params.status',
              lang: 'painless',
              params: {
                  status: 'failed'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('failed')

      #Attempt to terminate batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'terminate' endpoint failed, batch is in 'failed' state"
    end

    it 'Conflict: Batch that already has a terminated status' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

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
      response = @elastic.es_batch_update(TENANT_ID, @terminate_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @terminate_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('terminated')

      #Attempt to terminate batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'terminate' endpoint failed, batch is in 'terminated' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'terminate', nil)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing Authorization header')
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Unauthorized - No Roles' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_no_roles}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Must have hri_data_integrator role to update a batch')
    end

    it 'Unauthorized - Consumer Role Can Not Update Batch Status' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_consumer_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Must have hri_data_integrator role to update a batch')
    end

    it 'Unauthorized - Invalid Audience' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @terminate_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Terminate Batch Created With ID: #{@terminate_batch_id}")

      response = @hri_helper.hri_put_batch(TENANT_ID, @terminate_batch_id, 'terminate', nil, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

  end

  context 'PUT /tenants/{tenantId}/batches/{batchId}/action/fail' do

    before(:all) do
      @record_counts_and_message = {
          actualRecordCount: ACTUAL_RECORD_COUNT,
          failureMessage: FAILURE_MESSAGE,
          invalidRecordCount: INVALID_RECORD_COUNT
      }
      @batch_template = {
          name: @batch_name,
          dataType: DATA_TYPE,
          topic: BATCH_INPUT_TOPIC,
          invalidThreshold: INVALID_THRESHOLD,
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
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @failed_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Failed Batch Created With ID: #{@failed_batch_id}")

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
      response = @elastic.es_batch_update(TENANT_ID, @failed_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "completed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @failed_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('completed')

      #Set Batch to Failed
      response = @hri_helper.hri_put_batch(TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 200

      #Verify Batch Failed
      response = @hri_helper.hri_get_batch(TENANT_ID, @failed_batch_id, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 200
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['status']).to eql 'failed'
      expect(parsed_response['endDate']).to_not be_nil
      expect(parsed_response['invalidRecordCount']).to eq INVALID_RECORD_COUNT

      #Verify Kafka Message
      Timeout.timeout(KAFKA_TIMEOUT) do
        Logger.new(STDOUT).info("Waiting for a Kafka message with Batch ID: #{@failed_batch_id} and status: failed")
        @kafka_consumer.each_message do |message|
          parsed_message = JSON.parse(message.value)
          if parsed_message['id'] == @failed_batch_id && parsed_message['status'] == 'failed'
            @message_found = true
            expect(parsed_message['dataType']).to eql(DATA_TYPE)
            expect(parsed_message['id']).to eql(@failed_batch_id)
            expect(parsed_message['name']).to eql(@batch_name)
            expect(parsed_message['topic']).to eql(BATCH_INPUT_TOPIC)
            expect(parsed_message['invalidThreshold']).to eql(INVALID_THRESHOLD)
            expect(parsed_message['invalidRecordCount']).to eql(INVALID_RECORD_COUNT)
            expect(parsed_message['actualRecordCount']).to eql(ACTUAL_RECORD_COUNT)
            expect(parsed_message['failureMessage']).to eql(FAILURE_MESSAGE)
            expect(DateTime.parse(parsed_message['startDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(DateTime.parse(parsed_message['endDate']).strftime("%Y-%m-%d")).to eq(Date.today.strftime("%Y-%m-%d"))
            expect(parsed_message['metadata']['rspec1']).to eql('test1')
            expect(parsed_message['metadata']['rspec2']).to eql('test2')
            expect(parsed_message['metadata']['rspec3']['rspec3A']).to eql('test3A')
            expect(parsed_message['metadata']['rspec3']['rspec3B']).to eql('test3B')
            break
          end
        end
        expect(@message_found).to be true
      end
    end

    it 'Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Invalid Batch ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, INVALID_ID, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 404
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to include('document_missing_exception')
    end

    it 'Missing invalidRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', {actualRecordCount: ACTUAL_RECORD_COUNT, failureMessage: 'RSpec failure message'}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing required parameter(s): [invalidRecordCount]')
    end

    it 'Invalid invalidRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', {invalidRecordCount: "1", actualRecordCount: ACTUAL_RECORD_COUNT, failureMessage: 'RSpec failure message'}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Invalid parameter type(s): [invalidRecordCount must be a float64, got string instead.]')
    end

    it 'Missing actualRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', {invalidRecordCount: INVALID_RECORD_COUNT, failureMessage: 'RSpec failure message'}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing required parameter(s): [actualRecordCount]')
    end

    it 'Invalid actualRecordCount' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', {actualRecordCount: "1", invalidRecordCount: INVALID_RECORD_COUNT, failureMessage: 'RSpec failure message'}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Invalid parameter type(s): [actualRecordCount must be a float64, got string instead.]')
    end

    it 'Missing failureMessage' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', {actualRecordCount: ACTUAL_RECORD_COUNT, invalidRecordCount: INVALID_RECORD_COUNT}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing required parameter(s): [failureMessage]')
    end

    it 'Invalid failureMessage' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', {invalidRecordCount: INVALID_RECORD_COUNT, actualRecordCount: ACTUAL_RECORD_COUNT, failureMessage: 10}, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 400
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Invalid parameter type(s): [failureMessage must be a string, got float64 instead.]')
    end

    it 'Conflict: Batch with a status of terminated' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @failed_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Failed Batch Created With ID: #{@failed_batch_id}")

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
      response = @elastic.es_batch_update(TENANT_ID, @failed_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "terminated"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @failed_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('terminated')

      #Attempt to fail batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'fail' endpoint failed, batch is in 'terminated' state"
    end

    it 'Conflict: Batch that already has a failed status' do
      #Create Batch
      response = @hri_helper.hri_post_batch(TENANT_ID, @batch_template.to_json, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 201
      parsed_response = JSON.parse(response.body)
      @failed_batch_id = parsed_response['id']
      Logger.new(STDOUT).info("New Failed Batch Created With ID: #{@failed_batch_id}")

      #Update Batch to Failed Status
      update_batch_script = {
          script: {
              source: 'ctx._source.status = params.status',
              lang: 'painless',
              params: {
                  status: 'failed'
              }
          }
      }.to_json
      response = @elastic.es_batch_update(TENANT_ID, @failed_batch_id, update_batch_script)
      response.nil? ? (raise 'Elastic batch update did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['result']).to eql('updated')
      Logger.new(STDOUT).info('Batch status updated to "failed"')

      #Verify Batch Status Updated
      response = @elastic.es_get_batch(TENANT_ID, @failed_batch_id)
      response.nil? ? (raise 'Elastic get batch did not return a response') : (expect(response.code).to eq 200)
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['_source']['status']).to eql('failed')

      #Attempt to fail batch
      response = @hri_helper.hri_put_batch(TENANT_ID, @failed_batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_internal_role_only}"})
      expect(response.code).to eq 409
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql "The 'fail' endpoint failed, batch is in 'failed' state"
    end

    it 'Unauthorized - Missing Authorization' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message)
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Missing Authorization header')
    end

    it 'Unauthorized - Invalid Authorization' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_integrator_role_only}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql('Must have hri_internal role to mark a batch as failed')
    end

    it 'Unauthorized - Invalid Tenant ID' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_invalid_tenant}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

    it 'Unauthorized - Invalid Audience' do
      response = @hri_helper.hri_put_batch(TENANT_ID, @batch_id, 'fail', @record_counts_and_message, {'Authorization' => "Bearer #{@token_invalid_audience}"})
      expect(response.code).to eq 401
      parsed_response = JSON.parse(response.body)
      expect(parsed_response['errorDescription']).to eql("Unauthorized tenant access. Tenant '#{TENANT_ID}' is not included in the authorized scopes: .")
    end

  end

end