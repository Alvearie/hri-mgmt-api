# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

require 'dredd_hooks/methods'
require_relative '../env'
include DreddHooks::Methods

DEFAULT_TENANT_ID = 'provider1234'
TENANT_ID = "#{ENV['TRAVIS_BRANCH'].tr('.-', '')}".downcase
elastic = ElasticHelper.new

# uncomment to print out all the transaction names
before_all do |transactions|
  puts 'before all'
  @iam_token = IAMHelper.new.get_access_token
  # for transaction in transactions
  #   puts transaction['name']
  # end

  # make sure the tenant doesn't already exist
  elastic.delete_index(TENANT_ID)
end

before_each do |transaction|
  transaction['fullPath'].gsub!(DEFAULT_TENANT_ID, TENANT_ID)
end

before 'tenant > /tenants/{tenantId} > Create new tenant > 201 > application/json' do |transaction|
  puts 'before create tenant 201'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants/{tenantId} > Create new tenant > 401 > application/json' do |transaction|
  puts 'before create tenant 401'
  transaction['skip'] = false
end

before 'tenant > /tenants > Get a list of all tenantIds > 200 > application/json' do |transaction|
  puts 'before get tenants 200'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants > Get a list of all tenantIds > 401 > application/json' do |transaction|
  puts 'before get tenants 401'
  transaction['skip'] = false
end

before 'tenant > /tenants/{tenantId} > Get information on a specific elastic index > 200 > application/json' do |transaction|
  puts 'before get tenant 200'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants/{tenantId} > Get information on a specific elastic index > 401 > application/json' do |transaction|
  puts 'before get tenant 401'
  transaction['skip'] = false
end

before 'tenant > /tenants/{tenantId} > Get information on a specific elastic index > 404 > application/json' do |transaction|
  puts 'before get tenant 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(TENANT_ID, 'missingTenant')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'batch > /tenants/{tenantId}/batches > Create Batch > 201 > application/json' do |transaction|
  puts 'before create batch 201'
  transaction['skip'] = false
end

after 'batch > /tenants/{tenantId}/batches > Create Batch > 201 > application/json' do |transaction|
  puts 'after create batch 201'
  @batch_id = JSON.parse(transaction['real']['body'])['id']
end

before '/healthcheck > Perform a health check query of system availability > 200 > application/json' do |transaction|
  puts 'before heathcheck 200'
end

before 'stream > /tenants/{tenantId}/streams/{streamId} > Create new Stream for a Tenant > 201 > application/json' do |transaction|
  puts 'before create stream 201'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'stream > /tenants/{tenantId}/streams/{streamId} > Create new Stream for a Tenant > 400 > application/json' do |transaction|
  puts 'before create stream 400'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
  transaction['request']['body'] = '{bad json string"'
end

before 'stream > /tenants/{tenantId}/streams > Get all Streams for Tenant > 200 > application/json' do |transaction|
  puts 'before get streams 200'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'stream > /tenants/{tenantId}/streams > Get all Streams for Tenant > 404 > application/json' do |transaction|
  puts 'before get streams 404'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
  transaction['fullPath'].gsub!(TENANT_ID, 'invalid')
end

before 'stream > /tenants/{tenantId}/streams/{streamId} > Delete a Stream for a Tenant > 200 > application/json' do |transaction|
  puts 'before delete stream 200'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'stream > /tenants/{tenantId}/streams/{streamId} > Delete a Stream for a Tenant > 404 > application/json' do |transaction|
  puts 'before delete stream 404'
  transaction['skip'] = false
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
  transaction['fullPath'].gsub!(TENANT_ID, 'invalid')
end

before 'batch > /tenants/{tenantId}/batches > Create Batch > 400 > application/json' do |transaction|
  puts 'before create batch 400'
  transaction['skip'] = false
  transaction['request']['body'] = '{bad json string"'
end

before 'batch > /tenants/{tenantId}/batches > Create Batch > 404 > application/json' do |transaction|
  puts 'before create batch 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(TENANT_ID, 'missingTenant')
end

before 'batch > /tenants/{tenantId}/batches > Get Batches for Tenant > 404 > application/json' do |transaction|
  puts 'before get batches 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(TENANT_ID, 'missingTenant')
end

before 'batch > /tenants/{tenantId}/batches/{batchId} > Retrieve Metadata for Batch > 200 > application/json' do |transaction|
  puts 'before get batch 200'
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!('batch12345', @batch_id)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId} > Retrieve Metadata for Batch > 404 > application/json' do |transaction|
  puts 'before get batch 404'
  transaction['skip'] = false
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 200 > application/json' do |transaction|
  puts 'before sendComplete 200'
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    transaction['fullPath'].gsub!('batch12345', @batch_id)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 400 > application/json' do |transaction|
  puts 'before sendComplete 400'
  transaction['skip'] = false
  transaction['request']['body'] = '{bad json string"'
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 404 > application/json' do |transaction|
  puts 'before sendComplete 404'
  transaction['skip'] = false
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/sendComplete > Update Batch status to Send Complete > 409 > application/json' do |transaction|
  puts 'before sendComplete 409'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    elastic.es_batch_update(TENANT_ID, @batch_id[0..-6], '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "terminated"
          }
      }
    }')
    transaction['fullPath'].gsub!('batch12345', @batch_id)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 200 > application/json' do |transaction|
  puts 'before terminate 200'
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    elastic.es_batch_update(TENANT_ID, @batch_id[0..-6], '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "started"
          }
      }
    }')
    transaction['fullPath'].gsub!('batch12345', @batch_id)
  end
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 404 > application/json' do |transaction|
  puts 'before terminate 404'
  transaction['skip'] = false
end

before 'batch > /tenants/{tenantId}/batches/{batchId}/action/terminate > Terminate Batch > 409 > application/json' do |transaction|
  puts 'before terminate 409'
  transaction['skip'] = false
  if @batch_id.nil?
    transaction['fail'] = 'nil batch_id'
  else
    elastic.es_batch_update(TENANT_ID, @batch_id[0..-6], '
    {
      "script" : {
          "source": "ctx._source.status = params.status",
          "lang": "painless",
          "params" : {
              "status" : "completed"
          }
      }
    }')
    transaction['fullPath'].gsub!('batch12345', @batch_id)
  end
end

before 'tenant > /tenants/{tenantId} > Delete a specific tenant > 200 > application/json' do |transaction|
  puts 'before delete tenant 200'
  transaction['skip'] = false
  transaction['expected']['headers'].delete('Content-Type')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

before 'tenant > /tenants/{tenantId} > Delete a specific tenant > 401 > application/json' do |transaction|
  puts 'before delete tenant 401'
  transaction['skip'] = false
end

before 'tenant > /tenants/{tenantId} > Delete a specific tenant > 404 > application/json' do |transaction|
  puts 'before delete tenant 404'
  transaction['skip'] = false
  transaction['fullPath'].gsub!(TENANT_ID, 'invalid')
  transaction['request']['headers']['Authorization'] = "Bearer #{@iam_token}"
end

after_all do |transactions|
  puts 'after_all'
  # make sure the tenant index is deleted
  elastic.delete_index(TENANT_ID)
end