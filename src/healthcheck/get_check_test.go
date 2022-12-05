/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package healthcheck

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const requestId string = "testRequestId"

func TestGetCheck(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	requestId := "request_id_1"

	mt.Run("GetCheck", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		kafkaHealthChecker := fakeKafkaHealthChecker{}
		statusCode, err := GetCheck(requestId, kafkaHealthChecker)
		assert.NotNil(t, err)
		assert.Equal(t, statusCode, 503)

	})
}

type fakeKafkaHealthChecker struct {
	err error
}

func (fhc fakeKafkaHealthChecker) Check() error {
	return fhc.err
}

func (fhc fakeKafkaHealthChecker) Close() {
	return
}
