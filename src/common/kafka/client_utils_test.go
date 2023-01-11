package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTopicNames(t *testing.T) {
	tenantId := "tenant1"
	streamId := "dataIntegrator1.qualifier1"

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := CreateTopicNames(tenantId, streamId)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".in", inTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".notification", notificationTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".out", outTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".invalid", invalidTopicName)
}

func TestCreateTopicNamesNoQualifier(t *testing.T) {
	tenantId := "tenant1"
	streamId := "dataIntegrator1"

	inTopicName, notificationTopicName, outTopicName, invalidTopicName := CreateTopicNames(tenantId, streamId)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".in", inTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".notification", notificationTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".out", outTopicName)
	assert.Equal(t, "ingest."+tenantId+"."+streamId+".invalid", invalidTopicName)
}
