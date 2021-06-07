package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"log"
	"smart_intercom_api/pkg/config"
	"time"
)

func GenerateTokenForUser() (string, error) {
	serverConfig := config.GetConfig()

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Local().Add(serverConfig.TokenExpires).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(serverConfig.SecretKey)

	if err != nil {
		log.Fatal("Error in Generating key")
		return "", err
	}

	return tokenString, nil
}

func ParseTokenForUser(tokenStr string) error {
	serverConfig := config.GetConfig()

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return serverConfig.SecretKey, nil
		},
	)

	if err != nil {
		return err
	}

	_, ok := token.Claims.(*jwt.StandardClaims)

	if !ok {
		return errors.New("Couldn't parse claims")
	}

	return nil
}

func GenerateRefreshTokenForUser() (string, time.Time, error) {
	serverConfig := config.GetConfig()
	expiresTime := time.Now().Local().Add(serverConfig.RefreshTokenExpires)

	claims := &jwt.StandardClaims{
		ExpiresAt: expiresTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(serverConfig.SecretKey)

	if err != nil {
		log.Fatal("Error in Generating key")
		return "", expiresTime, err
	}

	return tokenString, expiresTime, nil
}

func ParseRefreshTokenForUser(tokenStr string) error {
	return ParseTokenForUser(tokenStr)
}

func GenerateTokenForPlugin(id string) (string, error) {
	serverConfig := config.GetConfig()

	claims := &jwt.StandardClaims{
		Id: id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(serverConfig.SecretKey)

	if err != nil {
		log.Fatal("Error in Generating key")
		return "", err
	}

	return tokenString, nil
}

func ParseTokenForPlugin(tokenStr string) (string, error) {
	serverConfig := config.GetConfig()

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return serverConfig.SecretKey, nil
		},
	)

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)

	if ok && token.Valid {
		id := claims.Id
		return id, nil
	}

	return "", errors.New("Couldn't parse claims")
}
