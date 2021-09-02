/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"strings"
)

const (
	inputSuffix           string = ".in"
	notificationSuffix    string = ".notification"
	statusErrPrefix       string = "Error extracting Batch Status: %s"
	statusErrMissingField string = "'status' field missing"
	statusErrInvalidValue string = "Invalid 'status' value: "
)

func EsDocToBatch(esDoc map[string]interface{}) map[string]interface{} {
	batch := esDoc["_source"].(map[string]interface{})
	batch[param.BatchId] = esDoc[esparam.EsDocId]
	batch = NormalizeBatchRecordCountValues(batch)
	return batch
}

func ExtractBatchStatus(batch interface{}) (status.BatchStatus, error) {
	var batchStatus = status.Unknown
	var err error = nil

	var batchMap = batch.(map[string]interface{})
	if statusField, ok := batchMap[param.Status]; ok {
		var statusStr = statusField.(string)
		var s = status.GetBatchStatus(statusStr)
		if s == status.Unknown {
			var msg = fmt.Sprintf(statusErrPrefix, statusErrInvalidValue+statusStr)
			err = errors.New(msg)
		} else {
			batchStatus = s
		}
	} else {
		var msg = fmt.Sprintf(statusErrPrefix, statusErrMissingField)
		err = errors.New(msg)
	}

	return batchStatus, err
}

// NormalizeBatchRecordCountValues ensures recordCount and expectedRecordCount do not exist alone on the provided batch.
// If one is missing, the unset property will be set to match the existing one.  If both or neither are missing, nothing is changed.
// Deprecated: Once the deprecated recordCount is completely removed, this function is no longer necessary.
func NormalizeBatchRecordCountValues(batch map[string]interface{}) map[string]interface{} {
	var recordCount, expectedRecordCount = batch[param.RecordCount], batch[param.ExpectedRecordCount]
	if recordCount != nil && expectedRecordCount == nil {
		batch[param.ExpectedRecordCount] = recordCount
	} else if recordCount == nil && expectedRecordCount != nil {
		batch[param.RecordCount] = expectedRecordCount
	}
	return batch
}

// InputTopicToNotificationTopic extracts a Notification topic from an inputTopic.
// Notification topic will be inferred from inputTopic using the following logic:
// If inputTopic ends with the ".in" suffix, then the suffix will be replaced with ".notification",
// If inputTopic does not end with the ".in" suffix, then ".notification" will just be appended to inputTopic
func InputTopicToNotificationTopic(inputTopic string) string {
	return strings.TrimSuffix(inputTopic, inputSuffix) + notificationSuffix
}
