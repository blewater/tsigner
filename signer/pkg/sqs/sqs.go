// Package sqs implements a messaging queue backed by AWS SQS
package sqs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"

	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/models"
)

const defaultRegion = "eu-west-2"

func getSecretRegion(cfg config.Configuration) string {
	if len(strings.TrimSpace(cfg.SQSRegion)) == 0 {
		return defaultRegion
	}

	return cfg.SQSRegion
}

func getLocalstackEndpoint(cfg config.Configuration) string {
	if (len(cfg.SQSLocalstackEndpoint)) == 0 {
		return "http://host.docker.internal:4566"
	}

	return cfg.SQSLocalstackEndpoint
}

// Client is an implementation of a sqs queue
type Client struct {
	sqsClient *sqs.Client

	writeSQSQueueURL string
}

// New creates an instance of a sqs queue implementation
func New(cfg config.Configuration) (*Client, error) {
	if len(cfg.SQSWriteQueueName) == 0 {
		return nil, errors.New("please provide the name of the queue to write to")
	}

	opts := []func(*awsConfig.LoadOptions) error{}

	opts = append(opts, awsConfig.WithRegion(getSecretRegion(cfg)))

	if config.IsLocal(cfg) {
		opts = append(opts, awsConfig.WithEndpointResolver(aws.EndpointResolverFunc(func(_, region string) (aws.Endpoint, error) { //nolint: staticcheck
			return aws.Endpoint{
				URL:               getLocalstackEndpoint(cfg),
				SigningRegion:     region,
				Source:            aws.EndpointSourceCustom,
				HostnameImmutable: true,
				PartitionID:       "aws",
			}, nil
		})))
	}

	conf, err := awsConfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	awstrace.AppendMiddleware(&conf)

	sqsClient := sqs.NewFromConfig(conf)

	writeQueueResult, err := sqsClient.GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.SQSWriteQueueName),
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		writeSQSQueueURL: *writeQueueResult.QueueUrl,
		sqsClient:        sqsClient,
	}, nil
}

// Close closes the underlying AWS connection
func (c *Client) Close() error { return nil }

// Add appends the item to the SQS queue
func (c *Client) Add(ctx context.Context, item models.SignedTXQueueItem) error {
	var b bytes.Buffer

	if err := json.NewEncoder(&b).Encode(item); err != nil {
		return err
	}

	_, err := c.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.writeSQSQueueURL),
		MessageBody: aws.String(b.String()),
	})

	return err
}
