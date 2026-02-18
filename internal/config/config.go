package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type ServerConfig struct {
	Workers   int `yaml:"workers"`
	QueueSize int `yaml:"queuesize"`
	Timeout   int `yaml:"timeout"`
}

type Config struct {
	Server    ServerConfig `yaml:"server"`
	InputFile string `yaml:"input_file"`
	OutputFile string `yaml:"output_file"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil,err
	}
	var cfg Config
	err = yaml.Unmarshal(data,&cfg)
	if err != nil {
		return nil,err
	}
	return &cfg,nil
}
