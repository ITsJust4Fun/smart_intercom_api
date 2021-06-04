package videos

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/pkg/config"
	"smart_intercom_api/pkg/random"
	"smart_intercom_api/pkg/subscriptions"
)

type Video struct {
	ID        string `json:"_id" bson:"_id"`
	Time      string `json:"time"`
	Link      string `json:"link"`
	Thumbnail string `json:"thumbnail"`
}

func videosCollection() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.3.14:27017"))

	if err != nil {
		log.Panic("Error when creating mongodb connection client", err)
	}

	collection := client.Database("smart_intercom_api").Collection("videos")
	err = client.Connect(ctx)

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	cancel()
	return collection
}

func (video *Video) InsertOne(input model.NewVideo) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := videosCollection()
	id, err := collection.InsertOne(ctx, input)

	if err != nil {
		cancel()
		log.Print("Error when inserting video", err)
		return err
	}

	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(video)

	if err != nil {
		cancel()
		log.Print("Error when finding the inserted video by its id", err)
		return err
	}

	cancel()
	return nil
}

func GetAll() ([]Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := videosCollection()
	result, err := collection.Find(ctx, bson.D{})

	if err != nil {
		cancel()
		log.Print("Error when finding video", err)
		return nil, err
	}

	defer func(result *mongo.Cursor, ctx context.Context) {
		err := result.Close(ctx)

		if err != nil {
			return
		}
	}(result, ctx)

	var videos []Video
	err = result.All(ctx, &videos)

	if err != nil {
		log.Print("Error when reading reports from cursor", err)
	}

	cancel()
	return videos, nil
}

func CreateVideoMutation(ctx context.Context, input model.NewVideo) (*model.Video, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	var video Video
	err := video.InsertOne(input)

	if err != nil {
		log.Print("Error when inserting video", err)
		return nil, err
	}

	result := model.Video(video)
	subscriptions.VideoUpdatedMutex.Lock()

	for _, observer := range subscriptions.VideoUpdatedObservers {
		observer <- &result
	}

	subscriptions.VideoUpdatedMutex.Unlock()

	return &result, nil
}

func Query(ctx context.Context) ([]*model.Video, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	allVideos, err := GetAll()

	if err != nil {
		log.Print("Error when getting videos", err)
	}

	var result []*model.Video

	for _, video := range allVideos {
		modelVideo := model.Video(video)
		result = append(result, &modelVideo)
	}

	return result, nil
}

func RemoveVideoMutation(ctx context.Context, input model.RemoveVideo) (*model.Video, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := videosCollection()

	id, _ := primitive.ObjectIDFromHex(input.ID)

	result, err := collection.DeleteOne(
		ctx,
		bson.M{"_id": id},
	)

	if err != nil {
		cancel()
		return nil, err
	}

	if result.DeletedCount != 1 {
		cancel()
		return nil, errors.New("can't find video to remove")
	}

	removedVideo := model.Video {
		ID: input.ID,
		Time: "removed",
		Link: "removed",
		Thumbnail: "removed",
	}

	subscriptions.VideoUpdatedMutex.Lock()

	for _, observer := range subscriptions.VideoUpdatedObservers {
		observer <- &removedVideo
	}

	subscriptions.VideoUpdatedMutex.Unlock()

	cancel()
	return &removedVideo, nil
}

func VideoUpdatedSubscription(ctx context.Context) (<-chan *model.Video, error) {
	id := random.String(8)
	videoEvent := make(chan *model.Video, 1)

	go func() {
		<-ctx.Done()
		subscriptions.VideoUpdatedMutex.Lock()
		delete(subscriptions.VideoUpdatedObservers, id)
		subscriptions.VideoUpdatedMutex.Unlock()
	}()

	subscriptions.VideoUpdatedMutex.Lock()
	subscriptions.VideoUpdatedObservers[id] = videoEvent
	subscriptions.VideoUpdatedMutex.Unlock()

	return videoEvent, nil
}
