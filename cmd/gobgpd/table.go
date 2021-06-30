package main

import (
	"context"
	"github.com/zhishi/R2RSimT/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
)

type MonitorMessage struct {
	Asn      uint32
	Action   string
	PathInfo interface{}
}

func UpdateRib(msg MonitorMessage)  {
	//
	//res, err := db.MongoCollection.InsertOne(context.Background(), bson.D{{"name", "pi"}, {"value", 3.14159}})
	//id := res.InsertedID
	db.MongoCollection.InsertOne(context.Background(), bson.D{{"asn", msg.Asn}, {"action", msg.Action},{"path_info",msg.PathInfo}})
}
