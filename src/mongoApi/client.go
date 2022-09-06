package mongoApi

import (
	"context"
	"fmt"

	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetMongoCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

func IndexFromTenantId(tenantId string) string {
	return tenantId + "-batches"
}

func LogAndBuildErrorDetail(requestId string, code int, logger logrus.FieldLogger,
	message string) *response.ErrorDetail {
	err := fmt.Errorf("%s: [%d]", message, code)
	logger.Errorln(err.Error())
	return response.NewErrorDetail(requestId, err.Error())
}

func DatabaseHealthCheck(client *mongo.Collection) string {
	command := bson.D{{"dbStats", 1}}
	var result bson.D
	client.Database().RunCommand(context.TODO(), command).Decode(&result)

	return fmt.Sprint(result[6].Value)
}
