package batches

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const docNotFoundMsg string = "The document for tenantId: %s with document (batch) ID: %s was not found"

func GetByBatchId(requestId string, batch model.GetByIdBatch, claims auth.HriAzClaims, client *mongo.Collection) (int, interface{}) {
	prefix := "batches/GetByBatchId"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenants Get By ID(Metadata)")
	// validate that caller has sufficient permissions
	if !claims.HasRole(auth.HriIntegrator) && !claims.HasRole(auth.HriConsumer) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "GetByBatchId")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	logger.Debugf("params_tenantID: %v, batchID: %v", batch.TenantId, batch.BatchId)
	var noAuthFlag = false
	return getByBatchId(requestId, batch, noAuthFlag, logger, &claims, client)

}

func GetByBatchIdNoAuth(requestId string, params model.GetByIdBatch,
	_ auth.HriAzClaims, client *mongo.Collection) (int, interface{}) {

	prefix := "batches/GetByBatchIdNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch GetById (No Auth)")

	var noAuthFlag = true
	return getByBatchId(requestId, params, noAuthFlag, logger, nil, client)
}
func getByBatchId(requestId string, batch model.GetByIdBatch,
	noAuthFlag bool, logger logrus.FieldLogger,
	claims *auth.HriAzClaims, client *mongo.Collection) (int, interface{}) {

	//Apending "-batches" to tenants id
	index := mongoApi.IndexFromTenantId(batch.TenantId)
	logger.Debugf("index: %v", index)

	var details []bson.M
	var getBatchMetaDataResponse model.GetBatchMetaDataResponse

	// Query Cosmos for information on the tenant
	projection := bson.D{
		{"batch", bson.D{
			{"$elemMatch", bson.D{
				{"batchid", batch.BatchId},
			}}}},
		{"_id", 0},
	}

	cursor, err := client.Find(
		context.TODO(),
		bson.D{
			{"tenantid", index},
		},
		options.Find().SetProjection(projection),
	)

	if err != nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Get batch by ID failed")
	}
	if err = cursor.All(context.TODO(), &details); err != nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Get batch by ID failed")
	}
	msg := fmt.Sprintf(msgDocNotFound, batch.TenantId, batch.BatchId)

	if details == nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, msg)
	}
	batchMap, ok := details[0]["batch"].(primitive.A)
	batchMapSlice := []interface{}(batchMap)

	if len(batchMapSlice) == 0 || !ok {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, msg)
	}

	getBatchMetaDataResponse = mapResponseBody(batchMapSlice)

	if !noAuthFlag {
		errDetailResponse := checkBatchAuth(requestId, claims, getBatchMetaDataResponse)
		if errDetailResponse != nil {
			return errDetailResponse.Code, errDetailResponse.Body
		}
	}

	return http.StatusOK, getBatchMetaDataResponse

}
func mapResponseBody(batchMapSlice []interface{}) (getBatchMetaData model.GetBatchMetaDataResponse) {
	var getBatchMetaDataResponse model.GetBatchMetaDataResponse

	mapResponseBody, _ := batchMapSlice[0].(primitive.M)

	mapResponse := map[string]interface{}(mapResponseBody)

	getBatchMetaDataResponse.DataType = mapResponse[param.HriDataType].(string)
	getBatchMetaDataResponse.Id = mapResponse[param.HriBatchId].(string)
	getBatchMetaDataResponse.IntegratorId = mapResponse[param.HriIntegratorId].(string)
	getBatchMetaDataResponse.InvalidThreshold = mapResponse[param.HriInvalidThreshold].(int32)
	getBatchMetaDataResponse.Metadata = mapResponse[param.HriMetadata].(primitive.M)
	getBatchMetaDataResponse.Name = mapResponse[param.HriName].(string)
	getBatchMetaDataResponse.StartDate = mapResponse[param.HriStartDate].(string)
	getBatchMetaDataResponse.Status = mapResponse[param.HriStatus].(string)
	getBatchMetaDataResponse.Topic = mapResponse[param.HriStatus].(string)
	return getBatchMetaDataResponse
}

// Data Integrators and Consumers can call this endpoint, but the behavior is slightly different. Consumers can see
// all Batches, but Data Integrators are only allowed to see Batches they created.
func checkBatchAuth(requestId string, claims *auth.HriAzClaims, resultBody model.GetBatchMetaDataResponse) *response.ErrorDetailResponse {
	if claims.HasRole(auth.HriConsumer) { //= Always Authorized
		return nil // return nil Error for Authorized
	}

	if claims.HasRole(auth.HriIntegrator) {
		integratorId := resultBody.IntegratorId
		//if claims.Subject from the token does NOT match the previously saved batch.IntegratorId, user NOT Authorized
		if claims.Subject != integratorId {
			errMsg := fmt.Sprintf(auth.MsgIntegratorSubClaimNoMatch, claims.Subject, integratorId)
			return response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errMsg)
		}
	}
	return nil //Default Return: we are Authorized  => nil error
}
