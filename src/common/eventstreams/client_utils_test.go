/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package eventstreams

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTopicNames(t *testing.T) {
	tenantId := "tenant1"
	streamId := "dataIntegrator1.qualifier1"

	inTopicName, notificationTopicName := CreateTopicNames(tenantId, streamId)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".in", inTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".notification", notificationTopicName)
}

func TestCreateTopicNamesNoQualifier(t *testing.T) {
	tenantId := "tenant1"
	streamId := "dataIntegrator1"

	inTopicName, notificationTopicName := CreateTopicNames(tenantId, streamId)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".in", inTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".notification", notificationTopicName)
}
