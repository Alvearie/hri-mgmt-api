package healthcheck

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const requestId string = "testRequestId"

func TestGetCheck(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.D{{"key", "value"}, {"key", "1"}, {"key", "1"}, {"health_status", "1"}, {"key", "1"}, {"key", "1"}}...))

		kafkaHealthChecker := fakeKafkaHealthChecker{}
		statusCode, err := GetCheck(requestId, kafkaHealthChecker)

		assert.Nil(t, err)
		assert.Equal(t, statusCode, http.StatusOK)

	})
	mt.Run("GetCheckStatusServiceUnavailable", func(mt *mtest.T) {

		mongoApi.HriCollection = mt.Coll
		kafkaHealthChecker := fakeKafkaHealthChecker{}
		statusCode, err := GetCheck(requestId, kafkaHealthChecker)
		assert.NotNil(t, err)
		assert.Equal(t, statusCode, http.StatusServiceUnavailable)

	})
	mt.Run("GetCheckStatusServiceUnavailable_KafkaClusterError", func(mt *mtest.T) {

		mongoApi.HriCollection = mt.Coll
		kafkaHealthChecker := fakeKafkaHealthChecker{
			err: errors.New("ResponseError contacting Kafka cluster: could not read partitions"),
		}

		statusCode, err := GetCheck(requestId, kafkaHealthChecker)
		assert.NotNil(t, err)
		assert.Equal(t, statusCode, http.StatusServiceUnavailable)

	})
	mt.Run("Kafka-health-check-returns-err", func(mt *mtest.T) {

		mongoApi.HriCollection = mt.Coll
		kafkaHealthChecker := fakeKafkaHealthChecker{
			err: errors.New("ResponseError contacting Kafka cluster: could not read partitions"),
		}

		statusCode, err := GetCheck(requestId, kafkaHealthChecker)
		assert.NotNil(t, err)
		assert.Equal(t, statusCode, http.StatusServiceUnavailable)

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
