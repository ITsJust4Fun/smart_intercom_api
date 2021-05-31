package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"smart_intercom_api/graph/generated"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/login"
	"smart_intercom_api/internal/videos"
	"smart_intercom_api/pkg/jwt"
)

func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	var authLogin login.Login
	authLogin.Password = input.Password
	correct := authLogin.Authenticate()

	if !correct {
		return "", &login.WrongPasswordError{}
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil{
		return "", err
	}

	return token, nil
}

func (r *mutationResolver) ChangePassword(ctx context.Context, input model.NewPassword) (string, error) {
	err := login.ChangePassword(input)

	if err != nil {
		return "", err
	}

	var authLogin login.Login
	authLogin.Password = input.PasswordNew
	correct := authLogin.Authenticate()

	if !correct {
		return "", errors.New("can't update password")
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil{
		return "", err
	}

	return token, nil
}

func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	err := jwt.ParseTokenForUser(input.Token)

	if err != nil {
		return "", fmt.Errorf("access denied")
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *mutationResolver) CreateVideo(ctx context.Context, input model.NewVideo) (*model.Video, error) {
	var video videos.Video
	err := video.InsertOne(input)

	if err != nil {
		log.Print("Error when inserting video", err)
		return nil, err
	}

	result := model.Video(video)

	return &result, nil
}

func (r *queryResolver) Videos(ctx context.Context) ([]*model.Video, error) {
	allVideos, err := videos.GetAll()

	if err != nil {
		log.Print("Error when getting users", err)
	}

	var result []*model.Video

	for _, video := range allVideos {
		modelVideo := model.Video(video)
		result = append(result, &modelVideo)
	}

	return result, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
