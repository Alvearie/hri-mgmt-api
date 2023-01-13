package tenants

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestDeleteOneTenant(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	requestId := "request_id_1"
	tenantId := "test-batches"

	tenantId2 := "test"
	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 1}})
		statusCode, response := DeleteTenant(requestId, tenantId)
		assert.NotNil(t, statusCode)
		assert.Equal(t, statusCode, 200)
		assert.Nil(t, response)
	})

	mt.Run("no document deleted", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 1}, {"acknowledged", true}, {"n", 0}})
		statusCode, response := DeleteTenant(requestId, tenantId2)
		assert.NotNil(t, statusCode)
		assert.Equal(t, statusCode, 404)
		assert.NotNil(t, response)
	})
}

/*func TestDelete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	requestId := "request_id_1"
	tenantId := "tenant123"

	elasticErrMsg := "elasticErrMsg"

	testCases := []struct {
		name              string
		validatorResponse map[string]interface{}
		transport         *test.FakeTransport
		expectedCode      int
		expectedBody      interface{}
	}{
		{
			name: "bad-response",
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf("Could not delete tenant [%s]: [500] elasticsearch client error: %s", tenantId, elasticErrMsg)),
		},
		{
			name: "good-request",
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseBody: fmt.Sprintf(`{"acknowledged":true,"shards_acknowledged":true,"index":"%s-batches"}`, tenantId),
				},
			),
			expectedCode: http.StatusOK,
			expectedBody: nil,
		},
	}

	for _, tc := range testCases {
		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			code, body := Delete(requestId, tenantId, client)
			if code != tc.expectedCode {
				t.Error(fmt.Sprintf("Incorrect HTTP code returned. Expected: [%v], actual: [%v]", tc.expectedCode, code))
			} else if !reflect.DeepEqual(tc.expectedBody, body) {
				t.Error(fmt.Sprintf("Incorrect HTTP response body returned. Expected: [%v], actual: [%v]", tc.expectedBody, body))
			}
			tc.transport.VerifyCalls()
		})
	}
}*/
