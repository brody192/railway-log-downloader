package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"main/internal/config/parser"
)

type config struct {
	DeploymentID  string `flag:"deployment" env:"RAILWAY_DEPLOYMENT_ID" usage:"deployment id to download logs for" validate:"uuid" required:"true"`
	Filter        string `flag:"filter" env:"RAILWAY_LOG_FILTER" usage:"filter to apply to logs"`
	OverwriteFile string `flag:"overwrite" env:"RAILWAY_OVERWRITE_FILE" usage:"overwrite existing logs file"`

	AccountToken string `env:"RAILWAY_ACCOUNT_TOKEN" usage:"railway account token" validate:"uuid" required:"true"`
}

var Railway = &config{}

func init() {
	// add help flag purely for the usage message
	flag.Bool("help", false, "Show help message")

	// Only parse and print usage if -help is present in arguments
	if checkForFlag("help") {
		parser.ParseFlags(Railway)

		flag.Usage()

		os.Exit(0)
	}

	errs := parser.ParseConfig(Railway)

	if len(errs) > 0 {
		fmt.Println("Error parsing config")
		fmt.Println(errors.Join(errs...))
		os.Exit(1)
	}
}
