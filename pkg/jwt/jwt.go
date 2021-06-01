package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"log"
	"time"
)

var (
	SecretKey = []byte("secret")
)

func GenerateTokenForUser() (string, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Local().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SecretKey)

	if err != nil {
		log.Fatal("Error in Generating key")
		return "", err
	}

	return tokenString, nil
}

func ParseTokenForUser(tokenStr string) error {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return SecretKey, nil
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
	expiresTime := time.Now().Local().Add(time.Hour * 24)

	claims := &jwt.StandardClaims{
		ExpiresAt: expiresTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SecretKey)

	if err != nil {
		log.Fatal("Error in Generating key")
		return "", expiresTime, err
	}

	return tokenString, expiresTime, nil
}

func ParseRefreshTokenForUser(tokenStr string) error {
	return ParseTokenForUser(tokenStr)
}
