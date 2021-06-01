package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"log"
	"smart_intercom_api/graph/generated"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/internal/login"
	"smart_intercom_api/internal/videos"
	"smart_intercom_api/pkg/jwt"
)

func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	var authLogin login.Login
	authLogin.Password = input.Password
	err := authLogin.Authenticate()

	if err != nil {
		return "", err
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil {
		return "", err
	}

	refreshToken, expiresTime, err := jwt.GenerateRefreshTokenForUser()

	if err != nil {
		return "", err
	}

	authLogin.RefreshToken = refreshToken
	cookieAccess := auth.GetCookieAccess(ctx)

	if cookieAccess == nil {
		return "", errors.New("can't get cookie")
	}

	cookieAccess.Token = refreshToken
	cookieAccess.Expires = expiresTime
	cookieAccess.SetToken()

	err = authLogin.ChangeRefreshToken()

	return token, err
}

func (r *mutationResolver) ChangePassword(ctx context.Context, input model.NewPassword) (string, error) {
	refresh, err := login.ChangePassword(input)

	if err != nil {
		return "", err
	}

	var authLogin login.Login
	authLogin.Password = input.PasswordNew
	err = authLogin.Authenticate()

	if err != nil {
		return "", err
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil {
		return "", err
	}

	cookieAccess := auth.GetCookieAccess(ctx)

	if cookieAccess == nil {
		return "", errors.New("can't get cookie")
	}

	cookieAccess.Token = refresh.Login.RefreshToken
	cookieAccess.Expires = refresh.Expires
	cookieAccess.SetToken()

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

func (r *queryResolver) RefreshToken(ctx context.Context) (string, error) {
	cookieAccess := auth.GetCookieAccess(ctx)

	if cookieAccess == nil {
		return "", errors.New("can't get cookie")
	}

	err := cookieAccess.GetToken()

	if err != nil {
		return "", err
	}

	loginData, err := login.GetLogin()

	if err != nil {
		return "", err
	}

	if loginData.RefreshToken != cookieAccess.Token {
		return "", errors.New("wrong refresh token")
	}

	err = jwt.ParseRefreshTokenForUser(cookieAccess.Token)

	if err != nil {
		return "", err
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil {
		return "", err
	}

	return token, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
