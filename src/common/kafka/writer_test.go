package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewConfluentWriter(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
		expErr error
	}{
		{

			name:   "successful construction",
			config: config.Config{AzKafkaBrokers: []string{"broker1", "broker2"}, AzKafkaProperties: config.StringMap{"message.max.bytes": "10000"}},
			expErr: nil,
		},
		{
			name:   "bad config",
			config: config.Config{AzKafkaBrokers: []string{"broker1", "broker2"}, AzKafkaProperties: config.StringMap{"message.max.bytes": "bad_value"}},
			expErr: fmt.Errorf("error constructing Kafka producer: %w",
				kafka.NewError(-186, "Invalid value for configuration property \"message.max.bytes\"", false)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWriterFromAzConfig(tt.config)

			assert.Equal(t, tt.expErr, err)
		})
	}
}

func TestConfluentKafkaWriter_Write(t *testing.T) {
	const (
		topic = "a.topic"
		key   = "a_unique_key"
	)
	goodValue := map[string]interface{}{"field1": "value", "field2": 10}
	var noError error = nil

	tests := []struct {
		name         string
		topic        string
		key          string
		value        map[string]interface{}
		produceErr   *error
		partitionErr *error
		expError     error
	}{
		{
			name:         "successfully send a message",
			topic:        topic,
			key:          key,
			value:        goodValue,
			produceErr:   &noError,
			partitionErr: &noError,
			expError:     nil,
		},
		{
			name:         "produce error",
			topic:        topic,
			key:          key,
			value:        goodValue,
			produceErr:   errPtr(errors.New("a message failure")),
			partitionErr: nil,
			expError:     fmt.Errorf("kafka producer error: %w", errors.New("a message failure")),
		},
		{
			name:         "TopicPartition error",
			topic:        topic,
			key:          key,
			value:        goodValue,
			produceErr:   &noError,
			partitionErr: errPtr(errors.New("topic does not exist")),
			expError:     fmt.Errorf("kafka producer error: %w", errors.New("topic does not exist")),
		},
	}

	controller := gomock.NewController(t)
	mockProducer := NewMockconfluentProducer(controller)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := confluentKafkaWriter{mockProducer}

			jsonVal, _ := json.Marshal(tt.value)

			// Only mock the producer methods if there is an expected response
			if tt.produceErr != nil {
				// expected message
				expMessage := &kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &tt.topic, Partition: kafka.PartitionAny},
					Key:            []byte(tt.key),
					Value:          jsonVal,
				}

				deliveryChan := make(chan kafka.Event)
				defer close(deliveryChan)

				mockProducer.EXPECT().
					Produce(expMessage, nil).
					DoAndReturn(func(message *kafka.Message, _ chan kafka.Event) interface{} {
						if tt.partitionErr != nil {
							message.TopicPartition.Error = *tt.partitionErr
							// this sends the message in another thread,
							// channels are blocking until both sender and receiver are ready
							go sendMessage(message, deliveryChan)
						}
						return *tt.produceErr
					})

				if tt.partitionErr != nil {
					mockProducer.EXPECT().Flush(1000)

					mockProducer.EXPECT().
						Events().
						Return(deliveryChan)
				}
			}

			err := writer.Write(tt.topic, tt.key, tt.value)

			assert.Equal(t, tt.expError, err)
		})
	}
}

func sendMessage(message *kafka.Message, channel chan kafka.Event) {
	channel <- message
}

// needed because you can't take the address of errors.New("")
func errPtr(err error) *error {
	return &err
}
