/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	kg "github.com/segmentio/kafka-go"
)

type PartitionReader interface {
	ReadPartitions(topics ...string) (partitions []kg.Partition, err error)
	Close() error
}

func CheckConnection(pr PartitionReader) (isAvailable bool, err error) {
	_, err = pr.ReadPartitions()
	if err != nil {
		return false, err
	}
	// We can't assume that there are any topics, so if `pr.ReadPartitions()` succeeds without an error
	// then we can assume Kafka is up and we can connect
	return true, nil
}
