/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"context"
	"encoding/json"
	kg "github.com/segmentio/kafka-go"
)

type Writer interface {
	Write(topic string, key string, val map[string]interface{}) error
}

type KafkaConnect struct {
	Brokers []string
	Dialer  *kg.Dialer
}

func (kc KafkaConnect) Write(topic string, key string, val map[string]interface{}) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// Configure batch size so that message is sent immediately, instead of waiting for 1-second timeout.
	// We also must set the TLS version to match example provided in Event Streams "Getting Started" guide
	writer := kg.NewWriter(kg.WriterConfig{
		Brokers:   kc.Brokers,
		Topic:     topic,
		Balancer:  kg.Murmur2Balancer{},
		BatchSize: 1,
		Dialer:    kc.Dialer,
	})

	defer writer.Close()

	// Context object is only useful in async mode, to cancel jobs, but we still must provide an instance
	return writer.WriteMessages(
		context.Background(),
		kg.Message{Key: []byte(key), Value: []byte(jsonVal)},
	)
}

func NewWriterFromParams(params map[string]interface{}) (Writer, error) {
	brokers, err := extractBrokers(params)
	if err != nil {
		return nil, err
	}

	dialer, err := CreateDialer(params)

	if err != nil {
		return nil, err
	}

	return KafkaConnect{
		Brokers: brokers,
		Dialer:  dialer,
	}, nil
}
