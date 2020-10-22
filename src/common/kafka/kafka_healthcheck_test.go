/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package kafka

import (
	"errors"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	kg "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

func TestReadKafkaPartition(t *testing.T) {

	var twoPartitions = test.GetFakeTwoPartitionSlice()
	zeroPartitions := []kg.Partition{}

	testCases := []struct {
		name        string
		reader      test.FakePartitionReader
		expected    bool
		expectedErr error
	}{
		{
			name: "success-multiple-partitions",
			reader: test.FakePartitionReader{
				T:          t,
				Partitions: twoPartitions,
				Err:        nil,
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "success-zero-partitions-slice",
			reader: test.FakePartitionReader{
				T:          t,
				Partitions: zeroPartitions,
				Err:        nil,
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "failure-partition-reader-error",
			reader: test.FakePartitionReader{
				T:          t,
				Partitions: nil,
				Err:        errors.New("Error contacting Kafka cluster: could not read partitions"),
			},
			expected:    false,
			expectedErr: errors.New("Error contacting Kafka cluster: could not read partitions"),
		}, {
			name: "success-nil-partitions",
			reader: test.FakePartitionReader{
				T:          t,
				Partitions: nil,
				Err:        nil,
			},
			expected:    true,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := CheckConnection(tc.reader)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Expected err to be %q but it was %q", tc.expectedErr, err)
			}
			assert.Equal(t, tc.expected, actual,
				"Expected Kafka HealthCheck (Conn) = "+strconv.FormatBool(tc.expected)+" for testCase: "+tc.name)

		})

	}

}
