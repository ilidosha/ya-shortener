package config

import (
	"github.com/jessevdk/go-flags"
	"os"
)

type Options struct {
	ServerAddress    string `short:"a" long:"address" description:"Server address" env:"SERVER_ADDRESS" default:"localhost:8080"`
	BaseURL          string `short:"b" long:"url" description:"Base URL for shortened URLs" env:"BASE_URL" default:"http://localhost:8080"`
	FileStore        string `short:"f" long:"file" description:"Base file storage path" env:"FILE_STORAGE_PATH" default:""`
	ConnectionString string `short:"d" long:"database" description:"Data base connection string" env:"DATABASE_DSN" default:""`
}

// ParseOptions parses the options from environment variables and command line arguments.
// Prioritizing env over command line arguments, and command line arguments over default values.
//
// Returns a pointer to Options struct and an error.
func ParseOptions() (*Options, error) {
	var opts, args Options

	// Define variables to store environment variables
	serverAddressEnv := os.Getenv("SERVER_ADDRESS")
	baseURLEnv := os.Getenv("BASE_URL")
	fileStoreEnv := os.Getenv("FILE_STORAGE_PATH")
	dataBaseEnv := os.Getenv("DATABASE_DSN")

	// Check if environment variables are set and assign them to the options
	if serverAddressEnv != "" {
		opts.ServerAddress = serverAddressEnv
	}
	if baseURLEnv != "" {
		opts.BaseURL = baseURLEnv
	}
	if fileStoreEnv != "" {
		opts.FileStore = fileStoreEnv
	}
	if dataBaseEnv != "" {
		opts.ConnectionString = dataBaseEnv
	}

	// Parse the command line arguments only if environment variables are not set
	if serverAddressEnv == "" || baseURLEnv == "" || fileStoreEnv == "" || dataBaseEnv == "" {
		parser := flags.NewParser(&args, flags.Default)
		_, err := parser.Parse()
		if err != nil {
			return nil, err
		}

		// Assign the command line arguments to the options if some of them not set already
		if serverAddressEnv == "" && args.ServerAddress != "" {
			opts.ServerAddress = args.ServerAddress
		}
		if baseURLEnv == "" && args.BaseURL != "" {
			opts.BaseURL = args.BaseURL
		}
		if fileStoreEnv == "" && args.FileStore != "" {
			opts.FileStore = args.FileStore
		}
		if dataBaseEnv == "" && args.ConnectionString != "" {
			opts.ConnectionString = args.ConnectionString
		}
	}

	return &opts, nil
}
