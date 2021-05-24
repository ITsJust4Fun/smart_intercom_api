package videos

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"smart_intercom_api/graph/model"
	"time"
)

type Video struct {
	ID   string `json:"_id" bson:"_id"`
	Time string `json:"time"`
	Link string `json:"link"`
}

func videosCollection() *mongo.Collection {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.3.14:27017"))

	if err != nil {
		log.Panic("Error when creating mongodb connection client", err)
	}

	collection := client.Database("smart_intercom_api").Collection("videos")
	err = client.Connect(ctx)

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	return collection
}

func (video *Video) InsertOne(input model.NewVideo) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := videosCollection()
	id, err := collection.InsertOne(ctx, input)

	if err != nil {
		log.Print("Error when inserting video", err)
		return err
	}

	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(video)

	if err != nil {
		log.Print("Error when finding the inserted video by its id", err)
		return err
	}

	return nil
}

func GetAll() ([]Video, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := videosCollection()
	result, err := collection.Find(ctx, bson.D{})

	if err != nil {
		log.Print("Error when finding user", err)
		return nil, err
	}

	defer result.Close(ctx)

	var videos []Video
	err = result.All(ctx, &videos)

	if err != nil {
		log.Print("Error when reading users from cursor", err)
	}

	return videos, nil
}
