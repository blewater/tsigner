// Package config holds all logic related to configuration
package config

import (
	"strings"

	"github.com/caarlos0/env/v9"
)

// Configuration holds important config values to start and configure the way
// the gRPC server starts up amongst others.
// This struct is filled up directly from environmental values
type Configuration struct {
	LogLevel    string `json:"level" env:"LOG_LEVEL"`
	Environment string `env:"ENVIRONMENT"`

	WalletsPostgresDSN      string `env:"WALLETS_POSTGRES_DSN"`
	TransactionsPostgresDSN string `env:"TRANSACTIONS_POSTGRES_DSN"`

	ServiceName string `env:"SERVICE_NAME"`

	SQSRegion             string `env:"SQS_REGION"`
	SQSWriteQueueName     string `env:"SQS_WRITE_QUEUE_NAME"`
	SQSLocalstackEndpoint string `env:"SQS_LOCALSTACK_ENDPOINT"`

	SepiorSecretName         string `env:"SEPIOR_SECRET_NAME"`
	SepiorSecretAWSRegion    string `env:"SEPIOR_SECRET_AWS_REGION"`
	SepiorLocalstackEndpoint string `env:"SEPIOR_LOCALSTACK_ENDPOINT"`

	MaraChainRPC string `env:"MARA_CHAIN_RPC"`
	MaraChainID  string `env:"MARA_CHAIN_ID"`
}

// IsLocal checks if the defined environment mode for the app is "local".
// This is important because some things like debugging and reflection api for
// the grpc server are exposed in dev/local mode only
// Must be turned off in production
func IsLocal(cfg Configuration) bool { return strings.ToUpper(cfg.Environment) == "LOCAL" }

// Load creates and fills up the Configuration struct with values from the
// environment from a .env file
func Load() (Configuration, error) {
	var cfg Configuration

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
