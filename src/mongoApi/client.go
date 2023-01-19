package mongoApi

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetMongoCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)

}

func GetTenantWithBatchesSuffix(tenantId string) string {
	return tenantId + "-batches"
}

func LogAndBuildErrorDetail(requestId string, code int, logger logrus.FieldLogger,
	message string) *response.ErrorDetail {
	err := fmt.Errorf("%s: [%d]", message, code)
	logger.Errorln(err.Error())
	return response.NewErrorDetail(requestId, err.Error())
}

func LogAndBuildErrorDetailWithoutStatusCode(requestId string, logger logrus.FieldLogger,
	message string) *response.ErrorDetail {
	err := fmt.Errorf("%s", message)
	logger.Errorln(err.Error())
	return response.NewErrorDetail(requestId, err.Error())
}

func HriDatabaseHealthCheck() (string, string, error) {
	command := bson.D{{Key: "dbStats", Value: 1}}
	var result bson.D
	err := HriCollection.Database().RunCommand(context.TODO(), command).Decode(&result)
	if result != nil {
		return fmt.Sprint(result[6].Value), fmt.Sprint(result[4].Value), err
	} else {
		return "", "", nil
	}
}

func TenantIdWithSuffix(tenant string) string {
	return strings.TrimSuffix(tenant, "-batches")
}
