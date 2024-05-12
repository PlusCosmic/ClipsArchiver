package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type StoreConfig struct {
	CacheStorePath string `json:"cacheStorePath"`
	StorePath      string `json:"storePath"`
}

type MatchHistoryConfig struct {
	AlsApiKey string `json:"apiKey"`
}

type DatabaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Name     string `json:"dbName"`
}

const configFileLoadError = "Error loading config file"
const inputPath = "/Uploads/"
const outputPath = "/Clips/"
const thumbnailsPath = "/Thumbnails/"
const resourcesPath = "/Resources/"
const storeConfigFile = "config.json"
const matchHistoryConfigFile = "apiConfig.json"
const dbConfigFile = "dbConfig.json"

var storeConfig *StoreConfig
var matchHistoryConfig *MatchHistoryConfig
var databaseConfig *DatabaseConfig
var configLoaded bool

func LoadConfig() {
	if CheckCreateConfigFiles() {
		fmt.Println("Config files were not found, they have been created now. Please populate and relaunch")
		os.Exit(0)
	}

	storeConfig = &StoreConfig{}
	matchHistoryConfig = &MatchHistoryConfig{}
	databaseConfig = &DatabaseConfig{}

	file, err := os.Open(storeConfigFile)
	if err != nil {
		log.Fatal(configFileLoadError)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(configFileLoadError)
		}
	}(file)
	fileBytes, err := io.ReadAll(file)
	err = json.Unmarshal(fileBytes, storeConfig)
	if err != nil {
		log.Fatal(configFileLoadError)
	}

	file, err = os.Open(matchHistoryConfigFile)
	if err != nil {
		log.Fatal(configFileLoadError)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(configFileLoadError)
		}
	}(file)
	fileBytes, err = io.ReadAll(file)
	err = json.Unmarshal(fileBytes, matchHistoryConfig)
	if err != nil {
		log.Fatal(configFileLoadError)
	}

	file, err = os.Open(dbConfigFile)
	if err != nil {
		log.Fatal(configFileLoadError)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(configFileLoadError)
		}
	}(file)
	fileBytes, err = io.ReadAll(file)
	err = json.Unmarshal(fileBytes, databaseConfig)
	if err != nil {
		log.Fatal(configFileLoadError)
	}
}

func CheckCreateConfigFiles() bool {
	anyFilesCreated := false
	if _, err := os.Stat(storeConfigFile); errors.Is(err, os.ErrNotExist) {
		anyFilesCreated = true
		file, err := os.Create(storeConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		newStoreConfig := StoreConfig{CacheStorePath: "", StorePath: ""}
		jsonBytes, err := json.Marshal(newStoreConfig)
		if err != nil {
			log.Fatal(err)
		}
		_, err = file.Write(jsonBytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(matchHistoryConfigFile); errors.Is(err, os.ErrNotExist) {
		anyFilesCreated = true
		file, err := os.Create(matchHistoryConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		newMatchHistoryConfig := MatchHistoryConfig{AlsApiKey: ""}
		jsonBytes, err := json.Marshal(newMatchHistoryConfig)
		if err != nil {
			log.Fatal(err)
		}
		_, err = file.Write(jsonBytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(dbConfigFile); errors.Is(err, os.ErrNotExist) {
		anyFilesCreated = true
		file, err := os.Create(dbConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		newDbConfig := DatabaseConfig{
			Username: "",
			Password: "",
			Address:  "",
			Name:     "",
		}
		jsonBytes, err := json.Marshal(newDbConfig)
		if err != nil {
			log.Fatal(err)
		}
		_, err = file.Write(jsonBytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	return anyFilesCreated
}

func GetInputPath() string {
	if !configLoaded {
		LoadConfig()
	}
	return storeConfig.CacheStorePath + inputPath
}

func GetOutputPath() string {
	if !configLoaded {
		LoadConfig()
	}
	return storeConfig.StorePath + outputPath
}

func GetThumbnailsPath() string {
	if !configLoaded {
		LoadConfig()
	}
	return storeConfig.CacheStorePath + thumbnailsPath
}

func GetResourcesPath() string {
	if !configLoaded {
		LoadConfig()
	}
	return storeConfig.CacheStorePath + resourcesPath
}

func GetApiKey() string {
	if !configLoaded {
		LoadConfig()
	}
	return matchHistoryConfig.AlsApiKey
}

func GetDatabaseInfo() *DatabaseConfig {
	if !configLoaded {
		LoadConfig()
	}
	return databaseConfig
}
