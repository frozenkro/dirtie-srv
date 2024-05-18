package provision

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Device struct {
	Id         string `bson:"_id,omitempty"`
	DeviceId   string `bson:"deviceId"`
	MacAddress string `bson:"macAddress"`
	Created    int32  `bson:"created"`
	LastComm   int32  `bson:"lastComm"`
}

func UpsertByMacAddress(macAddress string) *Device {
	client, ctx, dsc := connect()
	defer dsc()
	collection := client.Database("dirtie").Collection("devices")

	filter := bson.M{"macAddress": macAddress}
	update := bson.M{
		"$set": bson.M{"lastComm": time.Now().Unix()},
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	result := collection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil
		}
		panic(result.Err())
	}

	doc := Device{}
	err := result.Decode(&doc)
	if err != nil {
		panic(err)
	}

	return &doc
}

func connect() (*mongo.Client, context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb:27017"))
	if err != nil {
		panic(err)
	}

	return client, ctx, func() {
		disconnect(client, ctx, cancel)
	}
}

func disconnect(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}
