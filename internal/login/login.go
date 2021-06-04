package login

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/pkg/config"
	"smart_intercom_api/pkg/jwt"
	"time"
)

type Login struct {
	ID            string  `json:"_id" bson:"_id"`
	Password      string  `json:"password"`
	RefreshToken  string  `json:"refresh_token" bson:"refresh_token"`
}

type DataInsert struct {
	Password      string  `json:"password"`
	RefreshToken  string  `json:"refresh_token" bson:"refresh_token"`
}

type Refresh struct {
	Login *Login
	Expires time.Time
}

func loginsCollection() *mongo.Collection {
	serverConfig := config.GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), serverConfig.DatabaseTimeout)
	client, err := mongo.NewClient(options.Client().ApplyURI(serverConfig.DatabaseURI))

	if err != nil {
		log.Panic("Error when creating mongodb connection client", err)
	}

	collection := client.Database("smart_intercom_api").Collection("login")
	err = client.Connect(ctx)

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	cancel()
	return collection
}

func (login *Login) InsertOne(input model.Login) error {
	logins, err := GetAll()

	if len(logins) != 0 {
		log.Print("There is login", err)
		return errors.New("there is login")
	}

	login.Password, err = HashPassword(input.Password)

	if err != nil {
		return err
	}

	loginInsertData := DataInsert{
		Password: login.Password,
		RefreshToken: login.RefreshToken,
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := loginsCollection()
	id, err := collection.InsertOne(ctx, &loginInsertData)

	if err != nil {
		cancel()
		log.Print("Error when inserting login", err)
		return err
	}

	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(login)

	if err != nil {
		cancel()
		log.Print("Error when finding the inserted login by its id", err)
		return err
	}

	cancel()
	return nil
}

func GetAll() ([]Login, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := loginsCollection()
	result, err := collection.Find(ctx, bson.D{})

	if err != nil {
		cancel()
		log.Print("Error when finding user", err)
		return nil, err
	}

	defer func(result *mongo.Cursor, ctx context.Context) {
		err := result.Close(ctx)
		if err != nil {
			return
		}
	}(result, ctx)

	var logins []Login
	err = result.All(ctx, &logins)

	if err != nil {
		cancel()
		log.Print("Error when reading logins from cursor", err)
		return nil, err
	}

	cancel()
	return logins, nil
}

func GetLogin() (*Login, error) {
	logins, err := GetAll()

	if len(logins) != 1 {
		log.Print("Logins count != 1", err)
		return nil, errors.New("logins count != 1")
	}

	return &logins[0], nil
}

func ChangePassword(input model.NewPassword) (*Refresh, error) {
	logins, err := GetAll()

	if err != nil {
		return nil, err
	}

	if len(logins) == 0 {
		if input.PasswordOld != "" {
			return nil, &WrongPasswordError{}
		}

		var login Login
		var loginInput model.Login

		loginInput.Password = input.PasswordNew
		refreshToken, expiresTime, err := jwt.GenerateRefreshTokenForUser()

		if err != nil {
			return nil, err
		}

		login.RefreshToken = refreshToken
		err = login.InsertOne(loginInput)

		if err != nil {
			log.Print("Error when inserting login", err)
			return nil, err
		}

		refresh := Refresh{
			Login: &login,
			Expires: expiresTime,
		}

		return &refresh, nil
	} else if len(logins) == 1 {
		login := logins[0]

		if !CheckPasswordHash(input.PasswordOld, login.Password) {
			return nil, &WrongPasswordError{}
		}

		hashedPassword, err := HashPassword(input.PasswordNew)

		if err != nil {
			return nil, err
		}

		login.Password = hashedPassword
		refreshToken, expiresTime, err := jwt.GenerateRefreshTokenForUser()

		if err != nil {
			return nil, err
		}

		login.RefreshToken = refreshToken

		ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
		collection := loginsCollection()

		id, _ := primitive.ObjectIDFromHex(login.ID)

		_, err = collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.D{
				{"$set", bson.D{
					    {"password", login.Password},
					    {"refresh_token", login.RefreshToken},
				    },
				},
			},
		)

		if err != nil {
			cancel()
			return nil, err
		}

		refresh := Refresh{
			Login: &login,
			Expires: expiresTime,
		}

		cancel()
		return &refresh, nil
	}

	return nil, errors.New("logins count > 1")
}

func (login *Login) ChangeRefreshToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	collection := loginsCollection()

	id, _ := primitive.ObjectIDFromHex(login.ID)

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{{"$set", bson.D{{"refresh_token", login.RefreshToken}}}},
	)

	if err != nil {
		cancel()
		return err
	}

	cancel()
	return nil
}

func (login *Login) Authenticate() error {
	loginFromDB, err := GetLogin()

	if err != nil {
		return err
	}

	if !CheckPasswordHash(login.Password, loginFromDB.Password) {
		return &WrongPasswordError{}
	}

	login.ID = loginFromDB.ID
	login.Password = loginFromDB.Password
	login.RefreshToken = loginFromDB.RefreshToken

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func LoginMutation(ctx context.Context, input model.Login) (string, error) {
	var authLogin Login
	authLogin.Password = input.Password
	err := authLogin.Authenticate()

	if err != nil {
		return "", err
	}

	token, err := jwt.GenerateTokenForUser()

	if err != nil {
		return "", err
	}

	if !input.IsRemember {
		return "Bearer " + token, nil
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

	return "Bearer " + token, err
}

func ChangePasswordMutation(ctx context.Context, input model.NewPassword) (string, error) {
	refresh, err := ChangePassword(input)

	if err != nil {
		return "", err
	}

	var authLogin Login
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

	return "Bearer " + token, nil
}

func RefreshTokenQuery(ctx context.Context) (string, error) {
	cookieAccess := auth.GetCookieAccess(ctx)

	if cookieAccess == nil {
		return "", errors.New("can't get cookie")
	}

	err := cookieAccess.GetToken()

	if err != nil {
		return "", err
	}

	loginData, err := GetLogin()

	if err != nil {
		return "", err
	}

	if loginData.RefreshToken == "" {
		return "", errors.New("no refresh token")
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

	return "Bearer " + token, nil
}

func LogoutQuery(ctx context.Context) (string, error) {
	loginData, err := GetLogin()

	if err != nil {
		return "", err
	}

	if loginData.RefreshToken == "" {
		return "", errors.New("no refresh token")
	}

	loginData.RefreshToken = ""
	err = loginData.ChangeRefreshToken()

	if err != nil {
		return "", errors.New("can't remove token")
	}

	cookieAccess := auth.GetCookieAccess(ctx)

	if cookieAccess == nil {
		return "", errors.New("can't get cookie")
	}

	cookieAccess.DeleteToken()

	return "done", nil
}
