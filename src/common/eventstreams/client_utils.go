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
	OutSuffix          string = ".out"
	InvalidSuffix      string = ".invalid"
)

func CreateTopicNames(tenantId string, streamId string) (string, string, string, string) {
	// See documentation on HRI topic naming conventions: https://pages.github.ibm.com/wffh-hri/docs/admin.html#onboarding-new-data-integrators
	baseTopicName := strings.Join([]string{tenantId, streamId}, ".")
	inTopicName := TopicPrefix + baseTopicName + InSuffix
	notificationTopicName := TopicPrefix + baseTopicName + NotificationSuffix
	outTopicName := TopicPrefix + baseTopicName + OutSuffix
	invalidTopicName := TopicPrefix + baseTopicName + InvalidSuffix
	return inTopicName, notificationTopicName, outTopicName, invalidTopicName
}
