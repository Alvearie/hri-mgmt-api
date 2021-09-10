# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require 'dredd_hooks/methods'
require_relative '../env'
include DreddHooks::Methods

DEFAULT_TENANT_ID = 'provider1234'
TENANT_ID_TENANTS_STREAMS = "#{ENV['BRANCH_NAME'].tr('.-', '')}".downcase
TENANT_ID_BATCHES = 'test'

elastic = ElasticHelper.new

before_all do |transactions|
  puts 'before all'
  @iam_token = IAMHelper.new.get_access_token
  app_id_helper = AppIDHelper.new
  @token_all_roles = app_id_helper.get_access_token('hri_integration_tenant_test_integrator_consumer', 'tenant_test hri_data_integrator hri_consumer')
  @token_invalid_tenant = app_id_helper.get_access_token('hri_integration_tenant_test_invalid', 'tenant_test_invalid')

  # uncomment to print out all the transaction names
  # for transaction in transactions
  #   puts transaction['name']
  # end

  # make sure the tenant doesn't already exist
  elastic.delete_index(TENANT_ID_TENANTS_STREAMS)
end

# GET /healthcheck

before '/healthcheck > Perform a health check query of system availability > 200 > application/json' do |transaction|
  puts 'before heathcheck 200'
  transaction['skip'] = false
end

# POST /tenants/{tenant_id}

before 'tenant > /tenants/{tenantId} > Create new tenant > 201 > application/json' do |transaction|
  puts 'before create tenant 201'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants/{tenantId} > Create new tenant > 401 > application/json' do |transaction|
  puts 'before create tenant 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
end

# POST /tenants/{tenant_id}/streams/{integrator_id}

before 'stream > /tenants/{tenantId}/streams/{streamId} > Create new Stream for a Tenant > 201 > application/json' do |transaction|
  puts 'before create stream 201'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'stream > /tenants/{tenantId}/streams/{streamId} > Create new Stream for a Tenant > 400 > application/json' do |transaction|
  puts 'before create stream 400'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
  transaction['request']['body'] = '{bad json string"'
end

# DELETE /tenants/{tenant_id}/streams/{integrator_id}

before 'stream > /tenants/{tenantId}/streams/{streamId} > Delete a Stream for a Tenant > 200 > application/json' do |transaction|
  puts 'before delete stream 200'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'stream > /tenants/{tenantId}/streams/{streamId} > Delete a Stream for a Tenant > 404 > application/json' do |transaction|
  puts 'before delete stream 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, 'missingTenant')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

# DELETE /tenants/{tenant_id}

before 'tenant > /tenants/{tenantId} > Delete a specific tenant > 200 > application/json' do |transaction|
  puts 'before delete tenant 200'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['expected']['headers'].delete('Content-Type')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants/{tenantId} > Delete a specific tenant > 401 > application/json' do |transaction|
  puts 'before delete tenant 401'
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['skip'] = false
end

before 'tenant > /tenants/{tenantId} > Delete a specific tenant > 404 > application/json' do |transaction|
  puts 'before delete tenant 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, 'invalid')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

# GET /tenants/{tenant_id}/streams

before 'stream > /tenants/{tenantId}/streams > Get all Streams for Tenant > 200 > application/json' do |transaction|
  puts 'before get streams 200'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'stream > /tenants/{tenantId}/streams > Get all Streams for Tenant > 404 > application/json' do |transaction|
  puts 'before get streams 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, 'invalid')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

# GET /tenants

before 'tenant > /tenants > Get a list of all tenantIds > 200 > application/json' do |transaction|
  puts 'before get tenants 200'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants > Get a list of all tenantIds > 401 > application/json' do |transaction|
  puts 'before get tenants 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
end

# GET /tenants/{tenant_id}

before 'tenant > /tenants/{tenantId} > Get information on a specific elastic index > 200 > application/json' do |transaction|
  puts 'before get tenant 200'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants/{tenantId} > Get information on a specific elastic index > 401 > application/json' do |transaction|
  puts 'before get tenant 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_TENANTS_STREAMS)
end

before 'tenant > /tenants/{tenantId} > Get information on a specific elastic index > 404 > application/json' do |transaction|
  puts 'before get tenant 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, 'invalid')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

# GET /tenants/{tenant_id}/batches

before 'batch > /tenants/{tenantId}/batches > Get Batches for Tenant > 200 > application/json' do |transaction|
  puts 'before get batches 200'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
end

before 'batch > /tenants/{tenantId}/batches > Get Batches for Tenant > 401 > application/json' do |transaction|
  puts 'before get batches 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
end

# GET /tenants/{tenantId}/batches/{batchId}

before 'batch > /tenants/{tenantId}/batches/{batchId} > Retrieve Metadata for Batch > 200 > application/json' do |transaction|
  puts 'before get batch 200'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId} > Retrieve Metadata for Batch > 401 > application/json' do |transaction|
  puts 'before get batch 401'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId} > Retrieve Metadata for Batch > 404 > application/json' do |transaction|
  puts 'before get batch 404'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', 'INVALID')
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
  end
end

# POST /tenants/{tenant_id}/batches

before 'batch > /tenants/{tenantId}/batches > Create Batch > 201 > application/json' do |transaction|
  puts 'before create batch 201'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
end

after 'batch > /tenants/{tenantId}/batches > Create Batch > 201 > application/json' do |transaction|
  puts 'after create batch 201'
  transaction['skip'] = false
  parsed_body = JSON.parse transaction['real']['body']
  @batch_id = parsed_body['id']
end

before 'batch > /tenants/{tenantId}/batches > Create Batch > 400 > application/json' do |transaction|
  puts 'before create batch 400'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['body'] = '{bad json string"'
end

before 'batch > /tenants/{tenantId}/batches > Create Batch > 401 > application/json' do |transaction|
  puts 'before create batch 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
end

# PUT /tenants/{tenantId}/batches/{batchId}/action/sendComplete

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 200 > application/json' do |transaction|
  puts 'before sendComplete 200'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
    update_batch_script = {
        script: {
            source: 'ctx._source.status = params.status',
            lang: 'painless',
            params: {
                status: 'started'
            }
        }
    }.to_json
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, update_batch_script)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 400 > application/json' do |transaction|
  puts 'before sendComplete 400'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['body'] = '{bad json string"'
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 401 > application/json' do |transaction|
  puts 'before sendComplete 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 404 > application/json' do |transaction|
  puts 'before sendComplete 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 409 > application/json' do |transaction|
  puts 'before sendComplete 409'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "terminated"
          }
      }
    }')
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
  end
end

# PUT /tenants/{tenantId}/batches/{batchId}/action/processingComplete

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/processingComplete > Indicate the Batch has been processed (Internal) > 200 > application/json' do |transaction|
  puts 'before processingComplete 200'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!("batch12345", @batch_id)
    update_batch_script = {
        script: {
            source: 'ctx._source.status = params.status',
            lang: 'painless',
            params: {
                status: 'sendCompleted'
            }
        }
    }.to_json
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, update_batch_script)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/processingComplete > Indicate the Batch has been processed (Internal) > 400 > application/json' do |transaction|
  puts 'before processingComplete 400'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['body'] = '{bad json string"'
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/processingComplete > Indicate the Batch has been processed (Internal) > 401 > application/json' do |transaction|
  puts 'before processingComplete 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/processingComplete > Indicate the Batch has been processed (Internal) > 404 > application/json' do |transaction|
  puts 'before processingComplete 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/processingComplete > Indicate the Batch has been processed (Internal) > 409 > application/json' do |transaction|
  puts 'before processingComplete 409'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!("batch12345", @batch_id)
    update_batch_script = {
        script: {
            source: 'ctx._source.status = params.status',
            lang: 'painless',
            params: {
                status: 'completed'
            }
        }
    }.to_json
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, update_batch_script)
  end
end

# PUT /tenants/{tenantId}/batches/{batchId}/action/terminate

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 200 > application/json' do |transaction|
  puts 'before terminate 200'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "started"
          }
      }
    }')
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 401 > application/json' do |transaction|
  puts 'before terminate 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 404 > application/json' do |transaction|
  puts 'before terminate 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 409 > application/json' do |transaction|
  puts 'before terminate 409'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "completed"
          }
      }
    }')
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
    transaction['request']['headers']['Authorization'] = "Bearer #{@token_all_roles}"
  end
end

# PUT /tenants/{tenantId}/batches/{batchId}/action/fail

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/fail > Fail the Batch (Internal) > 200 > application/json' do |transaction|
  puts 'before fail 200'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "completed"
          }
      }
    }')
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!('batch12345', @batch_id)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/fail > Fail the Batch (Internal) > 400 > application/json' do |transaction|
  puts 'before fail 400'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['body'] = '{bad json string"'
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/fail > Fail the Batch (Internal) > 401 > application/json' do |transaction|
  puts 'before fail 401'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['request']['headers']['Authorization'] = "Bearer #{@token_invalid_tenant}"
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/fail > Fail the Batch (Internal) > 404 > application/json' do |transaction|
  puts 'before fail 404'
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
  transaction['skip'] = false
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/fail > Fail the Batch (Internal) > 409 > application/json' do |transaction|
  puts 'before fail 409'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID_BATCHES)
    transaction['fullPath'].gsub!("batch12345", @batch_id)
    update_batch_script = {
        script: {
            source: 'ctx._source.status = params.status',
            lang: 'painless',
            params: {
                status: 'failed'
            }
        }
    }.to_json
    elastic.es_batch_update(TENANT_ID_BATCHES, @batch_id, update_batch_script)
  end
end

after_all do |transactions|
  puts 'after_all'
  unless @batch_id.nil?
    elastic.es_delete_batch(TENANT_ID_BATCHES, @batch_id)
  end

  # make sure the tenant index is deleted
  elastic.delete_index(TENANT_ID_TENANTS_STREAMS)
end