# (C) Copyright IBM Corp. 2020
#
# SPDX-License-Identifier: Apache-2.0

class HRIHelper

  def initialize(url)
    @base_url = url
    @helper = Helper.new
    @iam_cloud_url = ENV['IAM_CLOUD_URL']
    @iam_token = IAMHelper.new.get_access_token
  end

  def hri_get_batches(tenant_id, query_params = nil, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}/batches"
    url += "?#{query_params}" unless query_params.nil?
    headers = { 'Accept' => 'application/json',
               'Content-Type' => 'application/json' }.merge(override_headers)
    @helper.rest_get(url, headers)
  end

  def hri_get_batch(tenant_id, batch_id, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}/batches/#{batch_id}"
    headers = { 'Accept' => 'application/json',
               'Content-Type' => 'application/json' }.merge(override_headers)
    @helper.rest_get(url, headers)
  end

  def hri_post_batch(tenant_id, request_body, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}/batches"
    headers = { 'Accept' => 'application/json',
               'Content-Type' => 'application/json' }.merge(override_headers)
    @helper.rest_post(url, request_body, headers)
  end

  def hri_put_batch(tenant_id, batch_id, action, record_count = {}, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}/batches/#{batch_id}/action/#{action}"
    headers = { 'Accept' => 'application/json',
               'Content-Type' => 'application/json' }.merge(override_headers)
    @helper.rest_put(url, record_count.to_json, headers)
  end

  def hri_post_tenant(tenant_id, request_body = nil, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}"}.merge(override_headers)
    @helper.rest_post(url, request_body, headers)
  end

  def hri_get_tenants(override_headers = {})
    url = "#{@base_url}/tenants"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}"}.merge(override_headers)
    @helper.rest_get(url, headers)
  end

  def hri_get_tenant(tenant_id, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}"}.merge(override_headers)
    @helper.rest_get(url, headers)
  end

  def hri_delete_tenant(tenant_id, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}"}.merge(override_headers)
    @helper.rest_delete(url, nil, headers, {})
  end

  def hri_get_tenant_streams(tenant_id, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}/streams"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}" }.merge(override_headers)
    @helper.rest_get(url, headers)
  end

  def hri_post_tenant_stream(tenant_id, integrator_id, request_body, override_headers = {})
    puts "POSTING TENANT STREAM WITH TENANT_ID: #{tenant_id} AND INTEGRATOR_ID: #{integrator_id}"
    url = "#{@base_url}/tenants/#{tenant_id}/streams/#{integrator_id}"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}" }.merge(override_headers)
    @helper.rest_post(url, request_body, headers, {})
  end

  def hri_delete_tenant_stream(tenant_id, integrator_id, override_headers = {})
    url = "#{@base_url}/tenants/#{tenant_id}/streams/#{integrator_id}"
    headers = { 'Accept' => 'application/json',
                'Content-Type' => 'application/json',
                'Authorization' => "Bearer #{@iam_token}" }.merge(override_headers)
    @helper.rest_delete(url, nil, headers, {})
  end

end