package kafka

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestCreateTopic(t *testing.T) {

	config := config.Config{KafkaProperties: config.StringMap{
		"security.protocol":                     "sasl_ssl",
		"sasl.mechanism":                        "PLAIN",
		"sasl.username":                         "token",
		"sasl.password":                         "vG4C7SnNaIebOqPUISvjOkMGZUlYAfjTjIeuNI5zm7Qe",
		"ssl.endpoint.identification.algorithm": "https",
	}}
	brokers := strings.Join([]string{
		"broker-1-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-5-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-3-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-4-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-2-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-0-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
	}, ",")

	topic := "djb.test.topic.in"
	numParts := 1
	replicationFactor := 3

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
	}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}
	a, err := kafka.NewAdminClient(kafkaConfig)
	assert.Nil(t, err)
	fmt.Printf("Made admin client.")

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx := context.Background()

	// Create topics on the cluster. Set admin options
	// to wait for the operation to finish (or at most 60s)
	/*maxDur, err := time.ParseDuration("10s")
	assert.Nil(t, err)*/
	fmt.Printf("Creating topics...")
	results, err := a.CreateTopics(
		ctx,
		//Multiple topics can be created simultaneously by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{
			{
				Topic:             "1_" + topic,
				NumPartitions:     numParts,
				ReplicationFactor: replicationFactor,
			}, {
				Topic:             "2_" + topic,
				NumPartitions:     numParts,
				ReplicationFactor: replicationFactor,
			},
		},
		kafka.SetAdminRequestTimeout(time.Second*5),
	)
	fmt.Printf("Returned.")

	if err != nil {
		fmt.Printf("Error returned by create call:\n")
		fmt.Printf(err.Error() + "\n")
		fmt.Printf("Code: %s", err.(kafka.Error).Code())

	} else {
		fmt.Printf("Create returned no error, but there may still be nested errors.")
	}

	// Print results
	for i, result := range results {

		fmt.Printf("Line %d:\n", i)
		fmt.Printf("Topic: %s\n", result.Topic)
		fmt.Printf("Error: %s\n", result.Error.Error())
		fmt.Printf("Error Code: %s\n\n", result.Error.Code())
	}

	a.Close()
}

func TestListTopics(t *testing.T) {

	config := config.Config{KafkaProperties: config.StringMap{
		"security.protocol":                     "sasl_ssl",
		"sasl.mechanism":                        "PLAIN",
		"sasl.username":                         "token",
		"sasl.password":                         "vG4C7SnNaIebOqPUISvjOkMGZUlYAfjTjIeuNI5zm7Q",
		"ssl.endpoint.identification.algorithm": "https",
	}}
	brokers := strings.Join([]string{
		"broker-1-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-5-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-3-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-4-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-2-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-0-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
	}, ",")

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
	}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}
	a, err := kafka.NewAdminClient(kafkaConfig)
	assert.Nil(t, err)
	fmt.Printf("Made admin client.\n\n")
	metadata, err := a.GetMetadata(nil, true, 10000)
	if metadata != nil {
		fmt.Printf("HELLO1\n\n")
	}
	if err != nil {
		fmt.Printf("Err returned by list\n\n")
		fmt.Printf("error: %s\n\n\n", err)
	}

	topics := metadata.Topics
	fmt.Printf("topics:\n\n")
	for key, value := range topics {
		fmt.Printf(key + "   ")
		fmt.Printf(value.Topic + "\n\n")
	}

	a.Close()
}

func TestDeleteTopic(t *testing.T) {

	config := config.Config{KafkaProperties: config.StringMap{
		"security.protocol":                     "sasl_ssl",
		"sasl.mechanism":                        "PLAIN",
		"sasl.username":                         "token",
		"sasl.password":                         "vG4C7SnNaIebOqPUISvjOkMGZUlYAfjTjIeuNI5zm7Q",
		"ssl.endpoint.identification.algorithm": "https",
	}}
	brokers := strings.Join([]string{
		"broker-1-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-5-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-3-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-4-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-2-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
		"broker-0-twvyj4m0kft5j0mh.kafka.svc01.us-east.eventstreams.cloud.ibm.com:9093",
	}, ",")

	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
	}
	for key, value := range config.KafkaProperties {
		kafkaConfig.SetKey(key, value)
	}
	a, err := kafka.NewAdminClient(kafkaConfig)
	assert.Nil(t, err)
	fmt.Printf("Made admin client.\n\n")

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//defer cancel()
	ctx := context.Background()

	fmt.Printf("Did defer cancel.")

	results, err := a.DeleteTopics(ctx, []string{"1_djb.test.topic.in", "2_djb.test.topic.in"}, kafka.SetAdminRequestTimeout(time.Second*5))

	if err != nil {
		fmt.Printf("Error returned by delete call:\n")
		fmt.Printf(err.Error() + "\n")
		fmt.Printf("Code: %s", err.(kafka.Error).Code())
	} else {
		fmt.Printf("Delete returned no error, but there may still be nested errors.")
	}

	// Print results
	for i, result := range results {
		fmt.Printf("Line %d:\n", i)
		fmt.Printf("Topic: %s\n", result.Topic)
		fmt.Printf("Error: %s\n", result.Error.Error())
		fmt.Printf("Error Code: %s\n\n", result.Error.Code())
	}

	a.Close()
}
