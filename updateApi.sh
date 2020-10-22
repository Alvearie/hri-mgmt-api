#!/bin/bash

ibmcloud fn api create /hri /tenants/{tenantId}/batches post hri_mgmt_api/create_batch --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/batches get hri_mgmt_api/get_batches --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/batches/{batchId} get hri_mgmt_api/get_batch_by_id --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/batches/{batchId}/action/sendComplete put hri_mgmt_api/send_complete --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/batches/{batchId}/action/terminate put hri_mgmt_api/terminate_batch --response-type http
ibmcloud fn api create /hri /healthcheck get hri_mgmt_api/healthcheck --response-type http
ibmcloud fn api create /hri /tenants/{tenantId} post hri_mgmt_api/create_tenant --response-type http
ibmcloud fn api create /hri /tenants/{tenantId} delete hri_mgmt_api/delete_tenant --response-type http
ibmcloud fn api create /hri /tenants get hri_mgmt_api/get_tenants --response-type http
ibmcloud fn api create /hri /tenants/{tenantId} get hri_mgmt_api/get_tenant_by_id --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/streams get hri_mgmt_api/get_streams --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/streams/{streamId} post hri_mgmt_api/create_stream --response-type http
ibmcloud fn api create /hri /tenants/{tenantId}/streams/{streamId} delete hri_mgmt_api/delete_stream --response-type http