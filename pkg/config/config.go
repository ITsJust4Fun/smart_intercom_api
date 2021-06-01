package config

import "time"

type Config struct {
	DatabaseURI          string
	DatabaseTimeout      time.Duration
	TokenExpires         time.Duration
	RefreshTokenExpires  time.Duration
	SecretKey            []byte
}

func GetConfig() Config {
	return Config{
		DatabaseURI: "mongodb://192.168.3.14:27017",
		DatabaseTimeout: 10 * time.Second,
		TokenExpires: 15 * time.Minute,
		RefreshTokenExpires: 24 * time.Hour,
		SecretKey: []byte("secret"),
	}
}
