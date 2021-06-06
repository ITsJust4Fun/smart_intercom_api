package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	DatabaseURI          string
	DiagnosticsProto     string
	DatabaseTimeout      time.Duration
	TokenExpires         time.Duration
	RefreshTokenExpires  time.Duration
	SecretKey            []byte
	IsLoaded             bool
}

type JsonData struct {
	DatabaseURI          string  `json:"database_uri"`
	DiagnosticsProto     string  `json:"diagnostics_proto"`
	DatabaseTimeout      int     `json:"database_timeout"`
	TokenExpires         int     `json:"token_expires"`
	RefreshTokenExpires  int     `json:"refresh_token_expires"`
	SecretKey            string  `json:"secret_key"`
}

var defaultConfig = Config{
	DatabaseURI: "mongodb://192.168.3.14:27017",
	DiagnosticsProto: "localhost:50051",
	DatabaseTimeout: 10 * time.Second,
	TokenExpires: 15 * time.Minute,
	RefreshTokenExpires: 24 * time.Hour,
	SecretKey: []byte("secret"),
	IsLoaded: false,
}

var loadedConfig = Config{IsLoaded: false}

func ReadConfigFile() {
	file, err := os.Open("config.json")

	if err != nil {
		return
	}

	jsonData := new(JsonData)

	decoder := json.NewDecoder(file)
	err = decoder.Decode(jsonData)

	if err != nil {
		return
	}

	loadedConfig.DatabaseURI = jsonData.DatabaseURI
	loadedConfig.DiagnosticsProto = jsonData.DiagnosticsProto
	loadedConfig.DatabaseTimeout = time.Duration(jsonData.DatabaseTimeout) * time.Second
	loadedConfig.TokenExpires = time.Duration(jsonData.TokenExpires) * time.Minute
	loadedConfig.RefreshTokenExpires = time.Duration(jsonData.RefreshTokenExpires) * time.Hour
	loadedConfig.IsLoaded = true
}

func GetConfig() Config {
	if !loadedConfig.IsLoaded {
		return defaultConfig
	}

	return loadedConfig
}
