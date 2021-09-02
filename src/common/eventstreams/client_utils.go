/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package eventstreams

import "strings"

const (
	TopicPrefix        string = "ingest."
	InSuffix           string = ".in"
	NotificationSuffix string = ".notification"
)

//See documentation on HRI topic naming conventions: https://alvearie.io/HRI/admin.html#onboarding-new-data-integrators
func CreateTopicNames(tenantId string, streamId string) (string, string) {
	baseTopicName := strings.Join([]string{tenantId, streamId}, ".")
	inTopicName := TopicPrefix + baseTopicName + InSuffix
	notificationTopicName := TopicPrefix + baseTopicName + NotificationSuffix
	return inTopicName, notificationTopicName
}
