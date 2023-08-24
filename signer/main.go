package main

import (
	"fmt"
	"os"
	"strings"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"

	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/datastore/postgres"
	"github.com/mara-labs/transactionsigner/pkg/sepior"
	"github.com/mara-labs/transactionsigner/pkg/sqs"
)

func getLevel(cfg config.Configuration) (log.Level, error) {
	switch strings.ToUpper(cfg.LogLevel) {
	case "INFO":
		return log.InfoLevel, nil
	case "DEBUG":
		return log.DebugLevel, nil
	case "WARN":
		return log.WarnLevel, nil
	case "ERROR":
		return log.ErrorLevel, nil
	case "FATAL":
		return log.FatalLevel, nil
	case "PANIC":
		return log.PanicLevel, nil
	case "TRACE":
		return log.TraceLevel, nil
	default:
		return log.InfoLevel, fmt.Errorf("unrecognized LOG_LEVEL value: %s", os.Getenv("LOG_LEVEL"))
	}
}

func main() {
	configValues, err := config.Load()
	if err != nil {
		log.WithError(err).Error("could not load configuration")
		os.Exit(1)
	}

	sqsClient, err := sqs.New(configValues)
	if err != nil {
		log.WithError(err).Error("could not connect to SQS")
		os.Exit(1)
	}

	signer, err := sepior.New(configValues)
	if err != nil {
		log.WithError(err).Error("could not initialize sepior client")
		os.Exit(1)
	}

	store, err := postgres.New(configValues)
	if err != nil {
		log.WithError(err).Error("could not connect to postgres")
		os.Exit(1)
	}

	handler := newHandler(configValues, signer, sqsClient, store)

	lambda.Start(ddlambda.WrapFunction(handler, nil))
}
