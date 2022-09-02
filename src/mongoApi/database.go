package mongoApi

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db *mongo.Database
)

// Connect ...
func ConnectFromConfig(config config.Config) error {
	prefix := "MongoClient"
	var logger = logwrapper.GetMyLogger("", prefix)
	logger.Debugln("Connecting to Mongo Client...")
	// Connect
	cl, err := mongo.NewClient(options.Client().ApplyURI(config.MongoDBUri))
	if err != nil {
		log.Println(err)
		log.Fatal("Cannot connect to database:")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = cl.Connect(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	db = cl.Database(config.MongoDBName)
	fmt.Println("Database Connected to", db.Name())
	return nil
}
