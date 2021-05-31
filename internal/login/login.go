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
	"time"
)

type Login struct {
	ID   string `json:"_id" bson:"_id"`
	Password string `json:"password"`
}

func loginsCollection() *mongo.Collection {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.3.14:27017"))

	if err != nil {
		log.Panic("Error when creating mongodb connection client", err)
	}

	collection := client.Database("smart_intercom_api").Collection("login")
	err = client.Connect(ctx)

	if err != nil {
		log.Panic("Error when connecting to mongodb", err)
	}

	return collection
}

func (login *Login) InsertOne(input model.Login) error {
	logins, err := GetAll()

	if len(logins) != 0 {
		log.Print("There is login", err)
		return errors.New("there is login")
	}

	input.Password, err = HashPassword(input.Password)

	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := loginsCollection()
	id, err := collection.InsertOne(ctx, input)

	if err != nil {
		log.Print("Error when inserting login", err)
		return err
	}

	err = collection.FindOne(ctx, bson.M{"_id": id.InsertedID}).Decode(login)

	if err != nil {
		log.Print("Error when finding the inserted login by its id", err)
		return err
	}

	return nil
}

func GetAll() ([]Login, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := loginsCollection()
	result, err := collection.Find(ctx, bson.D{})

	if err != nil {
		log.Print("Error when finding user", err)
		return nil, err
	}

	defer result.Close(ctx)

	var logins []Login
	err = result.All(ctx, &logins)

	if err != nil {
		log.Print("Error when reading logins from cursor", err)
	}

	return logins, nil
}

func GetPassword() (string, error) {
	logins, err := GetAll()

	if len(logins) != 1 {
		log.Print("Logins count != 1", err)
		return "", errors.New("logins count != 1")
	}

	return logins[0].Password, nil
}

func ChangePassword(input model.NewPassword) error {
	logins, err := GetAll()

	if err != nil {
		return err
	}

	if len(logins) == 0 {
		if input.PasswordOld != "" {
			return &WrongPasswordError{}
		}

		var login Login
		var loginInput model.Login

		loginInput.Password = input.PasswordNew

		err := login.InsertOne(loginInput)

		if err != nil {
			log.Print("Error when inserting login", err)
			return err
		}

		return nil
	} else if len(logins) == 1 {
		login := logins[0]

		if !CheckPasswordHash(input.PasswordOld, login.Password) {
			return &WrongPasswordError{}
		}

		hashedPassword, err := HashPassword(input.PasswordNew)

		if err != nil {
			return err
		}

		login.Password = hashedPassword

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		collection := loginsCollection()

		id, _ := primitive.ObjectIDFromHex(login.ID)

		_, err = collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.D{
				{"$set", bson.D{{"password", login.Password}}},
			},
		)

		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("logins count > 1")
}

func (login *Login) Authenticate() bool {
	hashedPassword, err := GetPassword()

	if err != nil {
		return false
	}

	return CheckPasswordHash(login.Password, hashedPassword)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
