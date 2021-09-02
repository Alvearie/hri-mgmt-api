/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"strings"
)

const (
	inputSuffix        string = ".in"
	notificationSuffix string = ".notification"
)

func EsDocToBatch(esDoc map[string]interface{}) map[string]interface{} {
	batch := esDoc["_source"].(map[string]interface{})
	batch[param.BatchId] = esDoc[esparam.EsDocId]
	return batch
}

// Notification topic will be inferred from inputTopic using the following logic:
// If inputTopic ends with the ".in" suffix, then the suffix will be replaced with ".notification",
// If inputTopic does not end with the ".in" suffix, then ".notification" will just be appended to inputTopic
func InputTopicToNotificationTopic(inputTopic string) string {
	return strings.TrimSuffix(inputTopic, inputSuffix) + notificationSuffix
}
