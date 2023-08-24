package sqs

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/models"
)

func TestClient_Add(t *testing.T) {
	queueName := os.Getenv("SQS_QUEUE_NAME")

	cfg := config.Configuration{
		Environment:           "local",
		SQSWriteQueueName:     queueName,
		SQSLocalstackEndpoint: os.Getenv("SQS_LOCALSTACK_ENDPOINT"),
		MaraChainID:           "123456",
	}

	client, err := New(cfg)

	require.NoError(t, err)

	require.NoError(t, client.Add(context.Background(), models.SignedTXQueueItem{}))
}
