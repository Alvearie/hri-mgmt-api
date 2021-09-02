/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

// These must match the values used in the manifest
const (
	TenantIndex = 3
	BatchIndex  = 5
	StreamIndex = 5

	BoundCreds       string = "__bx_creds"
	OpenWhiskHeaders string = "__ow_headers"
	OidcIssuer       string = "issuer"
	JwtAudienceId    string = "jwtAudienceId"

	BatchId      string = "id"
	DataType     string = "dataType"
	Metadata     string = "metadata"
	Name         string = "name"
	Status       string = "status"
	StartDate    string = "startDate"
	Topic        string = "topic"
	RecordCount  string = "recordCount"
	IntegratorId string = "integratorId"

	TenantId string = "tenantId"

	StreamId          string = "id"
	NumPartitions     string = "numPartitions"
	RetentionMs       string = "retentionMs"
	RetentionBytes    string = "retentionBytes"
	CleanupPolicy     string = "cleanupPolicy"
	SegmentMs         string = "segmentMs"
	SegmentBytes      string = "segmentBytes"
	SegmentIndexBytes string = "segmentIndexBytes"
)
