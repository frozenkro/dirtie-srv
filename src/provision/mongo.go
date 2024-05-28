package provision

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Device struct {
	Id         string `bson:"_id,omitempty"`
	MacAddress string `bson:"macAddress"`
	Created    int64  `bson:"created"`
	LastComm   int64  `bson:"lastComm"`
}

var (
	client *mongo.Client
)

func GetByMacAddress(macAddress string) (*Device, error) {
	collection := client.Database("dirtie").Collection("devices")

	filter := bson.M{"macAddress": macAddress}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		panic(result.Err())
	}

	doc := Device{}
	err := result.Decode(&doc)

	return &doc, err
}

func InsertDevice(device *Device) (string, error) {
	collection := client.Database("dirtie").Collection("devices")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	device.Created = time.Now().Unix()
	device.LastComm = device.Created

	result, err := collection.InsertOne(ctx, device)
	if err != nil {
		return "", err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", errors.New("mongo-driver InsertedID failed to assert as primitive.ObjectID")
}

func UpdateDeviceLastComm(id int) error {
	collection := client.Database("dirtie").Collection("devices")

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{"lastComm": time.Now().Unix()},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := collection.FindOneAndUpdate(ctx, filter, update)

	return result.Err()
}

func Connect() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri, ok := os.LookupEnv("MONGO_URI")
	if !ok {
		uri = "localhost:27017"
	}
	var connSt = fmt.Sprintf("mongodb://%s:%s@%s", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), uri)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connSt))
	if err != nil {
		panic(err)
	}

	return client
}

func Disconnect(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}
