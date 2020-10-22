/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

import "testing"

func TestTenantIdCheck(t *testing.T) {
	tests := []struct {
		name     string
		tenantId string
		errMsg   string
	}{
		{
			name:     "Good with alpha-numeric",
			tenantId: "tenant1234",
			errMsg:   "",
		},
		{
			name:     "Good with -",
			tenantId: "tenant-1234",
			errMsg:   "",
		},
		{
			name:     "Good with _",
			tenantId: "tenant_1234",
			errMsg:   "",
		},
		{
			name:     "Error with capital",
			tenantId: "tenantId1234",
			errMsg:   "TenantId: tenantId1234 must be lower-case alpha-numeric, '-', or '_'. 'I' is not allowed.",
		},
		{
			name:     "Error with $",
			tenantId: "tenant$1234",
			errMsg:   "TenantId: tenant$1234 must be lower-case alpha-numeric, '-', or '_'. '$' is not allowed.",
		},
		{
			name:     "Error with .",
			tenantId: "tenant.1234",
			errMsg:   "TenantId: tenant.1234 must be lower-case alpha-numeric, '-', or '_'. '.' is not allowed.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TenantIdCheck(tt.tenantId)
			if (err != nil && tt.errMsg != err.Error()) || (err == nil && tt.errMsg != "") {
				t.Errorf("TenantIdCheck() error = %v, expected %v", err, tt.errMsg)
			}
		})
	}
}

func TestStreamIdCheck(t *testing.T) {
	tests := []struct {
		name     string
		streamId string
		errMsg   string
	}{
		{
			name:     "Good with alpha-numeric",
			streamId: "stream1234",
			errMsg:   "",
		},
		{
			name:     "Good with -",
			streamId: "stream-1234",
			errMsg:   "",
		},
		{
			name:     "Good with _",
			streamId: "stream_1234",
			errMsg:   "",
		},
		{
			name:     "Error with capital",
			streamId: "streamId1234",
			errMsg:   "StreamId: streamId1234 must be lower-case alpha-numeric, '-', or '_', and no more than one '.'. 'I' is not allowed.",
		},
		{
			name:     "Error with ~",
			streamId: "stream~1234",
			errMsg:   "StreamId: stream~1234 must be lower-case alpha-numeric, '-', or '_', and no more than one '.'. '~' is not allowed.",
		},
		{
			name:     "Good with one.",
			streamId: "stream.1234",
			errMsg:   "",
		},
		{
			name:     "Error with 2 .",
			streamId: "stream.123.4",
			errMsg:   "StreamId: stream.123.4 must be lower-case alpha-numeric, '-', or '_', and no more than one '.'. 2 .'s are not allowed.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StreamIdCheck(tt.streamId)
			if (err != nil && tt.errMsg != err.Error()) || (err == nil && tt.errMsg != "") {
				t.Errorf("StreamIdCheck() error = %v, expected %v", err, tt.errMsg)
			}
		})
	}
}
