/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

const (
	batchId   string = "batch987"
	topicBase string = "batchTopic"
)

func TestCreateNewWriterFromConfig(t *testing.T) {
	cfg, err := config.GetConfig(test.FindConfigPath(t), nil)
	if err != nil {
		t.Error(err)
	}

	writer := NewWriterFromConfig(cfg)
	if writer == nil {
		t.Fatal("Returned Kafka client was nil")
	}
}

func TestWriterFromConfigJsonErrors(t *testing.T) {
	cfg, err := config.GetConfig(test.FindConfigPath(t), nil)

	kafkaWriter := NewWriterFromConfig(cfg)

	notificationTopic := topicBase + ".notification"
	value := make(chan int)
	invalidBatchJson := map[string]interface{}{"some str": value}
	err = kafkaWriter.Write(notificationTopic, batchId, invalidBatchJson)
	assert.NotNil(t, err)
	assert.Equal(t, "json: unsupported type: chan int", err.Error())

	badValue := math.Inf(1)
	invalidValueJson := map[string]interface{}{"some str": badValue}
	err2 := kafkaWriter.Write(notificationTopic, batchId, invalidValueJson)
	assert.NotNil(t, err2)
	assert.Equal(t, "json: unsupported value: +Inf", err2.Error())
}
