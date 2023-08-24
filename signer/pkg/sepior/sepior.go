// Package sepior implements functions to connect to Sepior and sign transactions
package sepior

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mara-labs/sepior/ethwallet"
	"github.com/mara-labs/sepior/session"
	log "github.com/sirupsen/logrus"
	"gitlab.com/sepior/go-tsm-sdk/sdk/tsm"

	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/models"
)

const (
	defaultSecretName = "sepior-user"
	defaultRegion     = "eu-west-2"
)

// Client implements a way to sign transactions with Sepior TSM nodes
type Client struct {
	tsmClient tsm.ECDSAClient
	chainID   *big.Int
}

// New returns a properly initialized Client to connect to Sepior
func New(configValues config.Configuration) (*Client, error) {
	chainID, ok := big.NewInt(0).SetString(configValues.MaraChainID, 10)
	if !ok {
		return nil, errors.New("invalid chain id")
	}

	sepiorConfig, err := retrieveSepiorConfig(configValues)
	if err != nil {
		return nil, err
	}

	cred, err := session.GetPersistedCredentialsFromReader(strings.NewReader(sepiorConfig))
	if err != nil {
		return nil, err
	}

	client, err := session.New(cred)
	if err != nil {
		return nil, err
	}

	return &Client{
		tsmClient: tsm.NewECDSAClient(client),
		chainID:   chainID,
	}, nil
}

// Sign signs the given transaction with
func (c *Client) Sign(_ context.Context, opts models.SignOptions) (
	*types.Transaction, error,
) {
	signer := types.NewEIP155Signer(c.chainID)

	return ethwallet.Sign(opts.TX,
		signer, c.tsmClient,
		opts.KeyID, opts.DerivationPath)
}

func getSecretRegion(cfg config.Configuration) string {
	if len(strings.TrimSpace(cfg.SepiorSecretAWSRegion)) == 0 {
		return defaultRegion
	}

	return cfg.SepiorSecretAWSRegion
}

func getSecretName(cfg config.Configuration) (string, error) {
	if len(strings.TrimSpace(cfg.SepiorSecretName)) == 0 {
		return "", errors.New("please set SEPIOR_SECRET_NAME in the environment as it is a required value")
	}

	return cfg.SepiorSecretName, nil
}

func getLocalstackEndpoint(cfg config.Configuration) string {
	if (len(cfg.SepiorLocalstackEndpoint)) == 0 {
		return "http://host.docker.internal:4566"
	}

	return cfg.SepiorLocalstackEndpoint
}

func retrieveSepiorConfig(cfg config.Configuration) (string, error) {
	sepiorUser, err := getSecretName(cfg)
	if err != nil {
		return "", err
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
		log.WithError(err).Error("could not set up AWS configuration")
		return "", err
	}

	secretsManager := secretsmanager.NewFromConfig(conf)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(sepiorUser),
		VersionStage: aws.String("AWSCURRENT"),
	}

	log.WithField("secret_name", sepiorUser).Debug("getting secret")

	result, err := secretsManager.GetSecretValue(context.Background(), input)
	if err != nil {
		log.WithError(err).WithField("secret_name", sepiorUser).
			WithField("region", getSecretRegion(cfg)).
			Error("could not fetch secret value")

		return "", err
	}

	return *result.SecretString, nil
}
