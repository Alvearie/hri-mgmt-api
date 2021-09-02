/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	kg "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	kafkaConnectionFailMsg string = "connecting to Kafka failed, error detail: %w"
	connectTimeout                = 30 * time.Second
	defaultNetwork         string = "tcp"
	defaultTLSVersion             = tls.VersionTLS12
)

type ContextDialer interface {
	DialContext(ctx context.Context, networkType string, address string) (*kg.Conn, error)
}

func CreateDialerFromConfig(config config.Config) *kg.Dialer {
	user := config.KafkaUsername
	password := config.KafkaPassword
	dialer := &kg.Dialer{
		SASLMechanism: plain.Mechanism{Username: user, Password: password},
		TLS: &tls.Config{
			MaxVersion: defaultTLSVersion,
		},
	}
	return dialer
}

func connectionFromConfig(config config.Config, dialer ContextDialer) (*kg.Conn, error) {
	prefix := "KafkaConnector"
	var logger = logwrapper.GetMyLogger("", prefix)
	brokers := config.KafkaBrokers

	logger.Debugln("Getting Kafka connection...")
	conn, err := getKafkaConn(dialer, brokers, logger)
	if err != nil {
		getConnectionError := fmt.Errorf(kafkaConnectionFailMsg, err)
		return nil, getConnectionError
	}

	return conn, err
}

func getKafkaConn(dialer ContextDialer, brokers []string, logger logrus.FieldLogger) (conn *kg.Conn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	for _, brokerAddr := range brokers {
		logger.Debugf("selected broker address: " + brokerAddr + "\n")
		if conn, err = dialer.DialContext(ctx, defaultNetwork, brokerAddr); err == nil {
			return conn, nil
		}
	}
	return
}

func ConnectionFromConfig(config config.Config) (*kg.Conn, error) {
	dialer := CreateDialerFromConfig(config)
	return connectionFromConfig(config, dialer)
}
