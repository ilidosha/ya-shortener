package config

import (
	"github.com/jessevdk/go-flags"
)

type Options struct {
	ServerAddress string `short:"a" long:"address" description:"Server address" env:"SERVER_ADDRESS" default:"localhost:8080"`
	BaseURL       string `short:"b" long:"url" description:"Base URL for shortened URLs" env:"BASE_URL" default:"http://localhost:8080"`
}

func ParseOptions() (*Options, error) {
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		return nil, err
	}
	return &opts, nil
}
