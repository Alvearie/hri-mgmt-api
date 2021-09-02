/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/golang/mock/gomock"
	kg "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testBroker1 string = "broker-0-porcupine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker2 string = "broker-1-porcupine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker3 string = "broker-2-porcupine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker4 string = "broker-3-porcupine.kafka.eventstreams.monkey.ibm.com:9093"
	testBroker5 string = "broker-4-porcupine.kafka.eventstreams.monkey.ibm.com:9093"
	badBroker   string = "bad-broker-address.monkey.ibm.com:9093"
)

func TestConnectorSuccess(t *testing.T) {
	config := config.Config{
		KafkaUsername: "token",
		KafkaPassword: "password",
		KafkaBrokers:  config.StringSlice{testBroker1, testBroker2, testBroker3},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	emptyConn := &kg.Conn{} //Upon Connection Success - we return a valid Conn obj.fakeConn := &kg.Conn{}
	mockDialer := test.NewMockContextDialer(controller)
	mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker1).
		Return(emptyConn, nil)

	rtnConn, err := connectionFromConfig(config, mockDialer)
	if assert.NoError(t, err) {
		assert.Equal(t, emptyConn, rtnConn)
	}
}

func TestConnector_ShouldIterateOverBrokerAddresses2Failures(t *testing.T) {
	config := config.Config{
		KafkaUsername: "token",
		KafkaPassword: "password",
		KafkaBrokers:  config.StringSlice{testBroker1, testBroker2, testBroker3, testBroker4, testBroker5},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()
	connError := errors.New("dial tcp 123.45.321.34:9093: i/o timeout ")

	emptyConn := &kg.Conn{} //Upon Connection Success - we return a valid Conn obj.
	mockDialer := test.NewMockContextDialer(controller)

	firstCall := mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker1).
		Return(nil, connError)

	secondCall := mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker2).
		Return(nil, connError).
		After(firstCall)

	mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, testBroker3).
		Return(emptyConn, nil).
		After(secondCall)

	rtnConn, err := connectionFromConfig(config, mockDialer)
	if assert.NoError(t, err) {
		assert.Equal(t, emptyConn, rtnConn)
	}
}

func TestConnectorConnectionFailure(t *testing.T) {
	config := config.Config{
		KafkaUsername: "token",
		KafkaPassword: "password",
		KafkaBrokers:  config.StringSlice{badBroker},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()
	connError := errors.New("dial tcp 123.45.321.34:9093: i/o timeout ")
	mockDialer := test.NewMockContextDialer(controller)

	mockDialer.
		EXPECT().
		DialContext(gomock.Any(), defaultNetwork, badBroker).
		Return(nil, connError)

	_, err := connectionFromConfig(config, mockDialer)
	expectedErr := fmt.Errorf(kafkaConnectionFailMsg, connError)
	assert.Equal(t, expectedErr, err)
}

func TestConnectionFromConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.Config
	}{
		{
			name: "Good params",
			config: config.Config{
				KafkaUsername: "token",
				KafkaPassword: "password",
				KafkaBrokers:  config.StringSlice{testBroker1, testBroker2, testBroker3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConnectionFromConfig(tt.config)
			assert.Error(t, err)
		})
	}
}
