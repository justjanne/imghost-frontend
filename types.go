package main

type Image struct {
	Id       string `json:"id"`
	MimeType string `json:"mime_type"`
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
	SIZE_FORMAT_COVER   = "cover"
	SIZE_FORMAT_CONTAIN = "contain"
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

type Config struct {
	Sizes         []SizeDefinition
	Quality       Quality
	SourceFolder  string
	TargetFolder  string
	Redis         RedisConfig
	ImageQueue    string
	ResultChannel string
}
