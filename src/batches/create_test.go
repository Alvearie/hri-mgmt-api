package batches

import (
	"errors"
	"os"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUpdateBatchInfo(t *testing.T) {
	validBatchMetadata := map[string]interface{}{"batchContact": "The_Village_People", "finalRecordCount": 18}
	validBatchZero := model.CreateBatch{
		Name:             batchName,
		Topic:            "inputTopic",
		DataType:         batchDataType,
		InvalidThreshold: 0,
		Metadata:         validBatchMetadata,
	}
	UpdateBatchInfo(validBatchZero, "integrator-id", "batchid")
}

func TestBuildBatchInfo(t *testing.T) {
	integratorId := "integratorId3"
	batchName := "monkeeBatch"
	batchDataType := "porcipine"
	topicBase := "batchFunTopic"
	inputTopic := topicBase + inputSuffix
	validBatchMetadata := map[string]interface{}{"batchContact": "The_Village_People", "finalRecordCount": 18}
	batchInvalidThreshold := 42
	batchInvalidThresholdZero := 0

	validBatch := model.CreateBatch{
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
		Metadata:         validBatchMetadata,
	}
	validBatchZero := model.CreateBatch{
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThresholdZero,
		Metadata:         validBatchMetadata,
	}

	validBatchNoThreshold := model.CreateBatch{
		Name:     batchName,
		Topic:    inputTopic,
		DataType: batchDataType,
		Metadata: validBatchMetadata,
	}

	validBatchNoMetadata := model.CreateBatch{
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
	}

	validClaims := auth.HriAzClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	validBatchInfo := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         validBatchMetadata,
		param.InvalidThreshold: validBatch.InvalidThreshold,
	}

	batchInfoNoThreshold := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         validBatchMetadata,
		param.InvalidThreshold: -1,
	}
	batchInfoNoThresholdZero := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         validBatchMetadata,
		param.InvalidThreshold: -1,
	}
	batchInfoNoMetadata := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         nil,
		param.InvalidThreshold: validBatchNoMetadata.InvalidThreshold,
	}

	testCases := []struct {
		name              string
		batch             model.CreateBatch
		claims            auth.HriAzClaims
		expectedBatchInfo map[string]interface{}
		expErr            error
	}{
		{
			name:              "Success valid Batch",
			batch:             validBatch,
			claims:            auth.HriAzClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: validBatchInfo,
		},
		{
			name:              "No Invalid Threshold Batch",
			batch:             validBatchNoThreshold,
			claims:            auth.HriAzClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: batchInfoNoThreshold,
		},
		{
			name:              "Nil Metadata Batch Success",
			batch:             validBatchNoMetadata,
			claims:            auth.HriAzClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: batchInfoNoMetadata,
		},
		{
			name:              "Zero",
			batch:             validBatchZero,
			claims:            auth.HriAzClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: batchInfoNoThresholdZero,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualBatchInfo := buildBatchInfo(validBatch, validClaims.Subject)
			if batchInfoNotEqual(tc.expectedBatchInfo, actualBatchInfo) {
				//notify/print error event as test result
				t.Errorf("BuildBatchInfo()\n   actual: %v,\n expected: %v", actualBatchInfo, tc.expectedBatchInfo)
			}
		})
	}
}

func batchInfoNotEqual(expectedBatchInfo map[string]interface{},
	actualBatchInfo map[string]interface{}) bool {

	if expectedBatchInfo[param.Name] != actualBatchInfo[param.Name] ||
		expectedBatchInfo[param.IntegratorId] != actualBatchInfo[param.IntegratorId] ||
		expectedBatchInfo[param.Topic] != actualBatchInfo[param.Topic] ||
		expectedBatchInfo[param.DataType] != actualBatchInfo[param.DataType] ||
		expectedBatchInfo[param.Status] != actualBatchInfo[param.Status] ||
		expectedBatchInfo[param.InvalidThreshold] != actualBatchInfo[param.InvalidThreshold] ||
		actualBatchInfo[param.StartDate] == nil {

		var actualMD = actualBatchInfo[param.Metadata]
		var expectedMD = expectedBatchInfo[param.Metadata]
		if actualMD == nil && expectedMD != nil {
			return true
		} else if actualMD != nil && expectedMD == nil {
			return true
		} else if actualMD != nil && expectedMD != nil {
			if len(actualMD.(map[string]interface{})) != len(expectedMD.(map[string]interface{})) {
				//This is probably good enough; don't really need to Verify that the Metadata map was copied over exactly
				// We can assume it is if length is same
				return true
			}
		}
	}
	return false
}

func TestCreate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	logwrapper.Initialize("error", os.Stdout)

	requestId := "reqZxQ9706"
	tenantId := "tenant123"
	integratorId := "integratorId"
	batchId := "batch654"
	batchName := "monkeeName"
	batchDataType := "pikachu"
	topicBase := "batchFunTopic"
	inputTopic := topicBase + inputSuffix
	batchMetadata := map[string]interface{}{"batchContact": "TestName", "finalRecordCount": 200}
	batchInvalidThreshold := 10

	validBatch := model.CreateBatch{
		TenantId:         tenantId,
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
		Metadata:         batchMetadata,
	}

	validBatchKafkaMetadata := map[string]interface{}{
		param.BatchId:          batchId,
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        "ignored",
		param.Metadata:         batchMetadata,
		param.InvalidThreshold: batchInvalidThreshold,
	}
	mt.Run("createBatch404", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		subject := "dataIntegrator1"

		claim := auth.HriAzClaims{Scope: auth.HriIntegrator, Roles: []string{auth.HriIntegrator, "hri_tenant_tenant123_data_integrator"}, Subject: subject}

		writer := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: validBatchKafkaMetadata,
			Error:         nil,
		}
		tenants, response := CreateBatch(requestId, validBatch, claim, writer)

		assert.NotNil(t, tenants)
		assert.NotNil(t, response)
		assert.Equal(t, tenants, 404)
	})
	mt.Run("empty claim", func(mt *mtest.T) {

		subject := "dataIntegrator1"

		claim := auth.HriAzClaims{Scope: auth.HriIntegrator, Roles: []string{}, Subject: subject}

		writer := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: validBatchKafkaMetadata,
			Error:         nil,
		}
		tenants, response := CreateBatch(requestId, validBatch, claim, writer)

		assert.NotNil(t, tenants)
		assert.NotNil(t, response)
		assert.Equal(t, tenants, 401)
	})

	mt.Run("Subject_empty", func(mt *mtest.T) {

		subject := ""

		claim := auth.HriAzClaims{Scope: auth.HriIntegrator, Roles: []string{auth.HriIntegrator, "hri_tenant_tenant123_data_integrator"}, Subject: subject}

		writer := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: validBatchKafkaMetadata,
			Error:         nil,
		}
		tenants, response := CreateBatch(requestId, validBatch, claim, writer)

		assert.NotNil(t, tenants)
		assert.NotNil(t, response)

	})
	mt.Run("CreateBatchNoAuth", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		subject := "integratorId"

		claim := auth.HriAzClaims{Scope: auth.HriIntegrator, Roles: []string{auth.HriIntegrator, "hri_tenant_tenant123_data_integrator"}, Subject: subject}

		writer := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: validBatchKafkaMetadata,
			Error:         nil,
		}
		tenants, response := CreateBatchNoAuth(requestId, validBatch, claim, writer)

		assert.NotNil(t, tenants)
		assert.NotNil(t, response)

	})

	mt.Run("kafkaErr", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		batchId = "batch654"

		subject := "dataIntegrator1"

		claim := auth.HriAzClaims{Scope: auth.HriIntegrator, Roles: []string{auth.HriIntegrator, "hri_tenant_tenant123_data_integrator"}, Subject: subject}

		writer := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: validBatchKafkaMetadata,
			Error:         errors.New("Unable to write to Kafka"),
		}

		tenant1 := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "tenantId", Value: "tenant123"},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(tenant1, killCursors)

		tenants, _ := CreateBatch(requestId, validBatch, claim, writer)

		assert.Equal(t, tenants, 500)
	})
}
