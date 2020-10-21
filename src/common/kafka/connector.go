/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	kg "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"log"
	"os"
	"time"
)

const (
	fieldBrokers      string        = "kafka_brokers_sasl"
	fieldUser         string        = "user"
	fieldPassword     string        = "password"
	emptyKafkaBrokers string        = "No Kafka network address provided in field %s from Kafka credentials"
	kafkaConnFailMsg  string        = "Connecting to Kafka Failed, error Detail: %v"
	connectTimeout    time.Duration = 30 * time.Second
	defaultNetwork    string        = "tcp"
	defaultTLSVersion               = tls.VersionTLS12
)

type ContextDialer interface {
	DialContext(ctx context.Context, networkType string, address string) (*kg.Conn, error)
}

func extractBrokers(params map[string]interface{}) ([]string, error) {
	creds, err := param.ExtractValues(params, param.BoundCreds, param.KafkaResourceId)
	if err != nil {
		return nil, err
	}

	brokersInterface, ok := creds[fieldBrokers].([]interface{})
	if !ok {
		return nil, errors.New(fmt.Sprintf(param.MissingKafkaFieldMsg, fieldBrokers))
	} else {
		brokers := make([]string, len(brokersInterface))
		for i, b := range brokersInterface {
			brokers[i] = b.(string)
		}
		return brokers, nil
	}
}

func CreateDialer(params map[string]interface{}) (*kg.Dialer, error) {
	user, err := param.ExtractString(params, fieldUser)
	if err != nil {
		return nil, err
	}

	password, err := param.ExtractString(params, fieldPassword)
	if err != nil {
		return nil, err
	}

	dialer := &kg.Dialer{
		SASLMechanism: plain.Mechanism{user, password},
		TLS: &tls.Config{
			MaxVersion: defaultTLSVersion,
		},
	}
	return dialer, nil
}

func ConnFromParams(params map[string]interface{}, dialer ContextDialer) (*kg.Conn, error) {
	logger := log.New(os.Stdout, "KafkaConnector: ", log.Llongfile)

	//Get the list of sasl broker addresses - use one of these
	brokers, err := extractBrokers(params)
	if err != nil {
		return nil, err
	} else if len(brokers) == 0 {
		return nil, errors.New(fmt.Sprintf(emptyKafkaBrokers, fieldBrokers))
	}

	logger.Printf("OK, lets get the Kafka Conn... \n")
	conn, err := getKafkaConn(dialer, brokers, logger)
	if err != nil { //do some extra err logging
		errMessage := fmt.Sprintf(kafkaConnFailMsg, err)
		logger.Println(errMessage)
		return nil, errors.New(errMessage)
	}

	return conn, err
}

func getKafkaConn(dialer ContextDialer, brokers []string, logger *log.Logger) (conn *kg.Conn, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	for _, brokerAddr := range brokers {
		logger.Printf("selected broker address: " + brokerAddr + "\n")
		if conn, err = dialer.DialContext(ctx, defaultNetwork, brokerAddr); err == nil {
			return conn, nil
		}
	}
	return
}
