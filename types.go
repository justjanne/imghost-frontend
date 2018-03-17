package main

import (
	"encoding/json"
	"os"
	"time"
)

type Image struct {
	Id           string `json:"id"`
	Title        string
	Description  string
	CreatedAt    time.Time
	OriginalName string
	MimeType     string `json:"mime_type"`
}

type Result struct {
	Id      string   `json:"id"`
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
}

type Size struct {
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Format string `json:"format"`
}

const (
	sizeFormatCover   = "cover"
	sizeFormatContain = "contain"
)

type Quality struct {
	CompressionQuality uint      `json:"compression_quality"`
	SamplingFactors    []float64 `json:"sampling_factors"`
}

type SizeDefinition struct {
	Size   Size   `json:"size"`
	Suffix string `json:"suffix"`
}

type RedisConfig struct {
	Address  string
	Password string
}

type DatabaseConfig struct {
	Format string
	Url    string
}

type Config struct {
	Sizes         []SizeDefinition
	Quality       Quality
	SourceFolder  string
	TargetFolder  string
	Redis         RedisConfig
	Database      DatabaseConfig
	ImageQueue    string
	ResultChannel string
}

func NewConfigFromEnv() Config {
	config := Config{}

	json.Unmarshal([]byte(os.Getenv("IK8R_SIZES")), &config.Sizes)
	json.Unmarshal([]byte(os.Getenv("IK8R_QUALITY")), &config.Quality)
	config.SourceFolder = os.Getenv("IK8R_SOURCE_FOLDER")
	config.TargetFolder = os.Getenv("IK8R_TARGET_FOLDER")
	config.Redis.Address = os.Getenv("IK8R_REDIS_ADDRESS")
	config.Redis.Password = os.Getenv("IK8R_REDIS_PASSWORD")
	config.ImageQueue = os.Getenv("IK8R_REDIS_IMAGE_QUEUE")
	config.ResultChannel = os.Getenv("IK8R_REDIS_RESULT_CHANNEL")
	config.Database.Format = os.Getenv("IK8R_DATABASE_TYPE")
	config.Database.Url = os.Getenv("IK8R_DATABASE_URL")

	return config
}
