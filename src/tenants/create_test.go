/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreateTenant(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	requestId := "request_id_1"
	tenantId := "test-batches"

	tenantId2 := "test"

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedTenant := model.CreateTenantRequest{
			//ID:       primitive.NewObjectID(),
			TenantId:     "test-batches",
			Docs_count:   "0",
			Docs_deleted: 0,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "tenants", mtest.FirstBatch, bson.D{
			//{"_id", expectedUser.ID},
			{"tenantId", expectedTenant.TenantId},
			{"docs.count", expectedTenant.Docs_count},
			{"docs.deleted", expectedTenant.Docs_deleted},
		}))

		statusCode, res := CreateTenant(requestId, tenantId)
		//respBody := map[string]interface{}{param.TenantId: tenantId}
		assert.NotNil(t, res)
		assert.Equal(t, statusCode, 201)

	})

	mt.Run("DuplicateTenant", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedTenant := model.CreateTenantRequest{
			//ID:       primitive.NewObjectID(),
			TenantId:     "test-batches",
			Docs_count:   "0",
			Docs_deleted: 0,
		}
		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{
				{"tenantId", expectedTenant.TenantId},
				{"docs.count", expectedTenant.Docs_count},
				{"docs.deleted", expectedTenant.Docs_deleted},
			}},
		})
		_, res := CreateTenant(requestId, tenantId2)
		//respBody := map[string]interface{}{param.TenantId: tenantId}
		assert.NotNil(t, res)

	})
}
