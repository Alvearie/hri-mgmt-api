/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package test

import (
	kg "github.com/segmentio/kafka-go"
	"testing"
)

const (
	Broker1            string = "broker-0-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	Broker2            string = "broker-1-porcypine.kafka.eventstreams.monkey.ibm.com:9093"
	defaultPartitionId int    = 5
	partition2Id       int    = 12
)

type FakePartitionReader struct {
	T          *testing.T
	Partitions []kg.Partition
	Err        error
}

func (f FakePartitionReader) ReadPartitions(topics ...string) (partitions []kg.Partition, err error) {
	if f.Err != nil {
		return []kg.Partition{}, f.Err
	}
	return f.Partitions, nil
}

func GetFakeTwoPartitionSlice() []kg.Partition {
	var leader = kg.Broker{
		Host: Broker1,
		Port: 1234,
		ID:   66,
		Rack: "",
	}
	var follower = kg.Broker{
		Host: Broker2,
		Port: 9093,
		ID:   25,
		Rack: "",
	}
	var replicas []kg.Broker
	replicas = append(replicas, leader)
	replicas = append(replicas, follower)
	var isr []kg.Broker
	return []kg.Partition{
		{Topic: "topic1",
			Leader:   leader,
			Replicas: replicas,
			Isr:      isr,
			ID:       defaultPartitionId},
		{
			Topic:    "topicZ03",
			Leader:   leader,
			Replicas: replicas,
			Isr:      isr,
			ID:       partition2Id},
	}
}
