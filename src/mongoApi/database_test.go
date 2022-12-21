package mongoApi

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestConnectFromConfig(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("success", func(mt *mtest.T) {
		cl = mt.Client
		HriCollection = mt.Coll
		db = mt.DB
		config := config.Config{
			MongoDBUri: "mongodb://hi",
		}
		err := ConnectFromConfig(config)
		assert.Nil(t, err)

	})
}

func TestConnectFromConfigBadURI(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("noConnToDB", func(mt *mtest.T) {
		cl = mt.Client
		HriCollection = mt.Coll
		db = mt.DB
		config := config.Config{
			MongoDBUri: "any",
		}
		err := ConnectFromConfig(config)
		assert.Error(t, err)

	})
}
