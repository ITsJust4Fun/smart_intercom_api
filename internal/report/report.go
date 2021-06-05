package report

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
)

type Report struct {
	ID        string `json:"_id" bson:"_id"`
	Level     int    `json:"level"`
	Time      string `json:"time"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	IsViewed  bool   `json:"is_viewed" bson:"is_viewed"`
}

type InsertReport struct {
	Level     int    `json:"level"`
	Time      string `json:"time"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	IsViewed  bool   `json:"is_viewed" bson:"is_viewed"`
}

func reportsCollection() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.3.14:27017"))

	if err != nil {
		log.Panic("Error when creating mongodb connection client", err)
	}

	collection := client.Database("smart_intercom_api").Collection("reports")
	err = client.Connect(ctx)

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	cancel()
	return collection
}

func (report *Report) InsertOne(input model.NewReport) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := reportsCollection()

	insertReport := InsertReport(input)

	id, err := collection.InsertOne(ctx, &insertReport)

	if err != nil {
		cancel()
		log.Print("Error when inserting report", err)
		return err
	}

	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(report)

	if err != nil {
		cancel()
		log.Print("Error when finding the inserted report by its id", err)
		return err
	}

	cancel()
	return nil
}

func GetAll() ([]Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := reportsCollection()
	result, err := collection.Find(ctx, bson.D{})

	if err != nil {
		cancel()
		log.Print("Error when finding reports", err)
		return nil, err
	}

	defer func(result *mongo.Cursor, ctx context.Context) {
		err := result.Close(ctx)

		if err != nil {
			return
		}

	}(result, ctx)

	var reports []Report
	err = result.All(ctx, &reports)

	if err != nil {
		log.Print("Error when reading reports from cursor", err)
	}

	cancel()
	return reports, nil
}

func CreateReportMutation(ctx context.Context, input model.NewReport) (*model.Report, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	var report Report
	err := report.InsertOne(input)

	if err != nil {
		log.Print("Error when inserting report", err)
		return nil, err
	}

	result := model.Report(report)

	return &result, nil
}

func ReportsQuery(ctx context.Context) ([]*model.Report, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	allReports, err := GetAll()

	if err != nil {
		log.Print("Error when getting reports", err)
	}

	var result []*model.Report

	for _, report := range allReports {
		modelReport := model.Report(report)
		result = append(result, &modelReport)
	}

	return result, nil
}

func RemoveReportMutation(ctx context.Context, input model.RemoveReport) (*model.Report, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := reportsCollection()

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
		return nil, errors.New("can't find report to remove")
	}

	removedReport := model.Report {
		ID: input.ID,
		Level: 0,
		Time: "removed",
		Title: "removed",
		Body: "removed",
		IsViewed: true,
	}

	cancel()
	return &removedReport, nil
}

func ViewReportMutation(ctx context.Context, input model.ViewReport) (*model.Report, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := reportsCollection()

	id, _ := primitive.ObjectIDFromHex(input.ID)

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{{"$set", bson.D{{"is_viewed", true}}}},
	)

	if err != nil {
		cancel()
		return nil, err
	}

	var report Report
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&report)

	if err != nil {
		cancel()
		log.Print("Error when finding the inserted report by its id", err)
		return nil, err
	}

	result := model.Report(report)
	cancel()
	return &result, nil
}

func UnviewedReportsCount(ctx context.Context) (int, error) {
	if !auth.GetLoginState(ctx) {
		return 0, errors.New("access denied")
	}

	reports, err := GetAll()

	if err != nil {
		return 0, err
	}

	unviewed := 0
	for _, report := range reports {
		if !report.IsViewed {
			unviewed += 1
		}
	}

	return unviewed, nil
}
