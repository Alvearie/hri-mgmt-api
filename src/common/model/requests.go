/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package model

type CreateBatch struct {
	TenantId         string                 `param:"tenantId" validate:"required,tenantid-validator"`
	Name             string                 `json:"name" validate:"required,injection-check-validator"`
	Topic            string                 `json:"topic" validate:"required,injection-check-validator"`
	DataType         string                 `json:"dataType" validate:"required,injection-check-validator"`
	InvalidThreshold int                    `json:"invalidThreshold"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type GetBatch struct {
	TenantId string  `param:"tenantId" validate:"required"`
	Name     *string `query:"name" validate:"omitempty,injection-check-validator"`
	Status   *string `query:"status" validate:"omitempty,injection-check-validator"`
	GteDate  *string `query:"gteDate" validate:"omitempty,injection-check-validator"`
	LteDate  *string `query:"lteDate" validate:"omitempty,injection-check-validator"`
	Size     *int    `query:"size"`
	From     *int    `query:"from"`
}

type GetByIdBatch struct {
	TenantId string `param:"tenantId" validate:"required"`
	BatchId  string `param:"id" validate:"required"`
}

type SendCompleteRequest struct {
	TenantId            string                 `param:"tenantId" validate:"required"`
	BatchId             string                 `param:"id" validate:"required"`
	ExpectedRecordCount *int                   `json:"expectedRecordCount" validate:"required_without=RecordCount,omitempty,min=0"`
	RecordCount         *int                   `json:"recordCount" validate:"required_without=ExpectedRecordCount,omitempty,min=0"`
	Metadata            map[string]interface{} `json:"metadata"`
	Validation          bool                   // not part of the incoming request
}

type TerminateRequest struct {
	TenantId string                 `param:"tenantId" validate:"required"`
	BatchId  string                 `param:"id" validate:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ProcessingCompleteRequest struct {
	TenantId           string `param:"tenantId" validate:"required"`
	BatchId            string `param:"id" validate:"required"`
	ActualRecordCount  *int   `json:"actualRecordCount" validate:"required,min=0"`
	InvalidRecordCount *int   `json:"invalidRecordCount" validate:"required,min=0"`
}

type FailRequest struct {
	ProcessingCompleteRequest
	FailureMessage string `json:"failureMessage" validate:"required"`
}

// CreateStreamsRequest The struct names have to be identical to the json names (with the first letter capitalized).
// error messages to return the same name.
type CreateStreamsRequest struct {
	TenantId          string  `param:"tenantId" validate:"required,tenantid-validator"`
	StreamId          string  `param:"id" validate:"required,streamid-validator"`
	NumPartitions     *int64  `json:"numPartitions" validate:"required,min=1,max=99"`
	RetentionMs       *int    `json:"retentionMs" validate:"required,min=3600000,max=2592000000"`
	CleanupPolicy     *string `json:"cleanupPolicy" validate:"omitempty,oneof=delete compact"`
	RetentionBytes    *int    `json:"retentionBytes" validate:"omitempty,min=10485760,max=1073741824"`
	SegmentMs         *int    `json:"segmentMs" validate:"omitempty,min=300000,max=2592000000"`
	SegmentBytes      *int    `json:"segmentBytes" validate:"omitempty,min=10485760,max=536870912"`
	SegmentIndexBytes *int    `json:"segmentIndexBytes" validate:"omitempty,min=102400,max=104857600"`
}

type GetStreamRequest struct {
	TenantId string `param:"tenantId" validate:"required"`
}

type DeleteStreamRequest struct {
	TenantId string `param:"tenantId" validate:"required"` // no tenant id validation
	StreamId string `param:"id" validate:"required,streamid-validator"`
}

type CreateTenant struct {
	TenantId string `param:"tenantId" validate:"required,tenantid-validator"`
}
