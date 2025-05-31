package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"main/internal/config/parser"
)

type ConfigString string

type config struct {
	DeploymentID  ConfigString `flag:"deployment" env:"RAILWAY_DEPLOYMENT_ID" usage:"deployment id to download logs for" validate:"uuid" required_one_of:"service_or_deployment"`
	ServiceID     ConfigString `flag:"service" env:"RAILWAY_SERVICE_ID" usage:"service id to download logs for (required if environment is provided)" validate:"uuid" required_one_of:"service_or_deployment" required_all:"environment_and_service"`
	EnvironmentID ConfigString `flag:"environment" env:"RAILWAY_ENVIRONMENT_ID" usage:"environment id to download logs for (required if service is provided)" validate:"uuid" required_all:"environment_and_service"`

	Filter        ConfigString `flag:"filter" env:"RAILWAY_LOG_FILTER" usage:"filter to apply to logs"`
	OverwriteFile ConfigString `flag:"overwrite" env:"RAILWAY_OVERWRITE_FILE" usage:"overwrite existing logs file" validate:"boolean"`
	Resume        ConfigString `flag:"resume" env:"RAILWAY_RESUME" usage:"resume downloading logs from the last downloaded log" validate:"boolean"`

	AccountToken ConfigString `env:"RAILWAY_ACCOUNT_TOKEN" usage:"railway account token" validate:"uuid" required:"true"`
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

// GetRequiredGroupValue returns the  value, and flag name for the field that is set in the specified required group
func (c *config) GetRequiredGroupValue(groupName string) (flagName, value string) {
	return parser.GetRequiredGroupValue(c, groupName)
}

func (c *ConfigString) String() string {
	return *(*string)(c)
}

func (c *ConfigString) Bool() bool {
	b, _ := strconv.ParseBool(*(*string)(c))

	return b
}
