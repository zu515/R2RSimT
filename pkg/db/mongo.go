package db

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net"
	"time"
	"context"
)

func initMongo() (*mongo.Client,error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	credential := options.Credential{
		Username: FileConfig.Mongo.User,
		Password: FileConfig.Mongo.Password,
	}
	Mongo, err := mongo.Connect(ctx,
		options.Client().ApplyURI("mongodb://"+ net.JoinHostPort(FileConfig.Mongo.Host, FileConfig.Mongo.Port)).SetAuth(credential),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "initMongo",
		}).Errorf("err:%v", err)
		return Mongo,err
	}
	defer func() {
		if err = Mongo.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	err = Mongo.Ping(ctx, readpref.Primary())
	if err != nil{
		log.WithFields(log.Fields{
			"Topic": "initMongo Ping",
		}).Errorf("err:%v", err)
		return Mongo, err
	}
	MongoCollection = Mongo.Database("admin").Collection("BGP")
	return Mongo, nil
}