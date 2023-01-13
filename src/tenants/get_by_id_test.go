package tenants

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetTenantById(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	requestId := "request_id_1"
	tenantId := "test-batches"
	mt.Run("tenantNotFound", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedTenant := model.TenatGetResponse{
			Health:      "red",
			Status:      "close",
			Index:       "test-batches",
			Uuid:        primitive.NewObjectID(),
			Size:        "270336",
			DocsCount:   "0",
			DocsDeleted: "0",
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "getTenantById", mtest.FirstBatch, bson.D{
			{Key: "health", Value: expectedTenant.Health},
			{"_id", expectedTenant.Uuid},
			{"docs.count", expectedTenant.DocsCount},
			{"docs.deleted", expectedTenant.DocsDeleted},
			{"status", expectedTenant.Status},
			{"tenantId", expectedTenant.Index},
			{"size", expectedTenant.Size},
		}))
		_, err := GetTenantById(requestId, tenantId)
		assert.NotNil(t, err)
		//assert.Equal(t, &expectedTenant, tenantResponse)
	})

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedUser := model.GetTenantDetail{
			Uuid:         primitive.NewObjectID(),
			TenantId:     "test-batches",
			Docs_count:   "10",
			Docs_deleted: 0,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"_id", expectedUser.Uuid},
			{"tenantId", expectedUser.TenantId},
			{"docs_count", expectedUser.Docs_count},
			{"docs_deleted", expectedUser.Docs_deleted},
		}))
		statusCode, response := GetTenantById(requestId, tenantId)
		assert.NotNil(t, response)
		assert.Equal(t, statusCode, 200)
	})

	mt.Run("success-EmptyDocCount", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedUser := model.GetTenantDetail{
			Uuid:         primitive.NewObjectID(),
			TenantId:     "test-batches",
			Docs_count:   "",
			Docs_deleted: 0,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"_id", expectedUser.Uuid},
			{"tenantId", expectedUser.TenantId},
			{"docs_count", expectedUser.Docs_count},
			{"docs_deleted", expectedUser.Docs_deleted},
		}))
		statusCode, response := GetTenantById(requestId, tenantId)
		assert.NotNil(t, response)
		assert.Equal(t, statusCode, 200)
	})
}

/*
func TestGetById(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	validTenantId := "pgten901"
	invalidTenantId := "TENANT_NO_EXISTO"
	requestId := "reqVOQ5589"
	elasticErrMsg := "elasticErrMsg"

	getByIdResponse := map[string]interface{}{
		"index":          validTenantId + "-batches",
		"health":         "green",
		"status":         "open",
		"uuid":           "vTBmiZwhRcatGw4qAQqdRQ",
		"pri":            "1",
		"rep":            "1",
		"docs.count":     "75",
		"docs.deleted":   "2",
		"store.size":     "108.8kb",
		"pri.store.size": "54.4kb",
	}

	testCases := []struct {
		name         string
		requestId    string
		tenantId     string
		transport    *test.FakeTransport
		expectedCode int
		expectedBody interface{}
	}{
		{
			name:      "bad-response",
			requestId: requestId,
			tenantId:  validTenantId,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/_cat/indices/%s-batches", validTenantId),
				test.ElasticCall{
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewErrorDetail(requestId,
				fmt.Sprintf("Could not retrieve tenant '"+validTenantId+"': [500] elasticsearch client error: %s", elasticErrMsg),
			),
		},
		{
			name:      "Tenant not found",
			requestId: requestId,
			tenantId:  invalidTenantId,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/_cat/indices/%s-batches", invalidTenantId),
				test.ElasticCall{
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody: `
						{
  							"error" : {
    						"root_cause" : [
      							{
									"type" : "index_not_found_exception",
        							"reason" : "no such index",
        							"resource.type" : "index_or_alias",
        							"resource.id" : "TENANT_NO_EXISTO-batches",
        							"index_uuid" : "_na_",
        							"index" : "TENANT_NO_EXISTO-batches"
      							}
    						],
    						"type" : "index_not_found_exception",
    						"reason" : "no such index",
							"resource.type" : "index_or_alias",
    						"resource.id" : "TENANT_NO_EXISTO-batches",
							"index_uuid" : "_na_",
    						"index" : "TENANT_NO_EXISTO-batches"
  						},
  						"status" : 404
					}
				`,
				},
			),
			expectedCode: http.StatusNotFound,
			expectedBody: response.NewErrorDetail(requestId,
				"Tenant: TENANT_NO_EXISTO not found: [404] index_not_found_exception: no such index"),
		},
		{
			name:      "success-case",
			requestId: requestId,
			tenantId:  validTenantId,
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/indices/"+validTenantId+"-batches",
				test.ElasticCall{
					ResponseBody: `
					[
						{
							"health" : "green",
    						"status" : "open",
    						"index" : "pgten901-batches",
							"uuid" : "vTBmiZwhRcatGw4qAQqdRQ",
    						"pri" : "1",
    						"rep" : "1",
   							"docs.count" : "75",
    						"docs.deleted" : "2",
							"store.size" : "108.8kb",
    						"pri.store.size" : "54.4kb"
  						}
					]`,
				},
			),
			expectedCode: http.StatusOK,
			expectedBody: getByIdResponse,
		},
	}

	for _, tc := range testCases {
		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			actualCode, actualBody := GetById(tc.requestId, tc.tenantId, client)

			tc.transport.VerifyCalls()
			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
				t.Errorf("Tenant-GetById()\n   actual: %v,%v\n expected: %v,%v",
					actualCode, actualBody, tc.expectedCode, tc.expectedBody)
			}
		})
	}
}*/
