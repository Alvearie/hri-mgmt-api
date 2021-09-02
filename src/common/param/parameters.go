/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

const (
	BoundCreds string = "__bx_creds"

	BatchId             string = "id"
	TenantId            string = "tenantId"
	StreamId            string = "id"
	DataType            string = "dataType"
	Metadata            string = "metadata"
	Name                string = "name"
	IntegratorId        string = "integratorId"
	Status              string = "status"
	StartDate           string = "startDate"
	GteDate             string = "gteDate"
	LteDate             string = "lteDate"
	Topic               string = "topic"
	RecordCount         string = "recordCount" // deprecated
	ExpectedRecordCount string = "expectedRecordCount"
	ActualRecordCount   string = "actualRecordCount"
	InvalidThreshold    string = "invalidThreshold"
	InvalidRecordCount  string = "invalidRecordCount"
	FailureMessage      string = "failureMessage"

	Size string = "size"
	From string = "from"
)
