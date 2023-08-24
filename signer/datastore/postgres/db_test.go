//go:build integration
// +build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/docker/go-connections/nat"
	testfixtures "github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var databaseNames = []string{"walletstest", "transactionstest"}

type PostgresDatabaseTestSuite struct {
	suite.Suite

	walletsDBContainer, transactionDBContainer testcontainers.Container
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(PostgresDatabaseTestSuite))
}

func (p *PostgresDatabaseTestSuite) SetupSuite() {
	for _, v := range databaseNames {

		containerReq := testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
			Env: map[string]string{
				"POSTGRES_DB":       v,
				"POSTGRES_PASSWORD": "maratest",
				"POSTGRES_USER":     "maratest",
			},
		}

		container, err := testcontainers.GenericContainer(
			context.Background(),
			testcontainers.GenericContainerRequest{
				ContainerRequest: containerReq,
				Started:          true,
			})

		require.NoError(p.T(), err)

		if v == "walletstest" {
			p.walletsDBContainer = container
			continue
		}

		p.transactionDBContainer = container
	}
}

func (p *PostgresDatabaseTestSuite) TearDownSuite() {
	err := p.walletsDBContainer.Terminate(context.Background())
	require.NoError(p.T(), err)

	err = p.transactionDBContainer.Terminate(context.Background())
	require.NoError(p.T(), err)
}

func (p *PostgresDatabaseTestSuite) SetupTest() {
	for _, v := range databaseNames {

		db, err := sql.Open("postgres", p.getDSN(v))
		require.NoError(p.T(), err)

		err = db.Ping()
		require.NoError(p.T(), err)

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(p.T(), err)

		extraPath := "trx"
		if v == "walletstest" {
			extraPath = "wallets"
		}

		migrator, err := migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s", fmt.Sprintf("testdata/migrations/%s", extraPath)),
			"postgres", driver)

		require.NoError(p.T(), err)

		err = migrator.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			require.NoError(p.T(), err)
		}

		fixtures, err := testfixtures.New(
			testfixtures.Database(db),
			testfixtures.Dialect("postgres"),
			testfixtures.Directory(fmt.Sprintf("testdata/fixtures/%s", extraPath)),
		)

		require.NoError(p.T(), err)

		require.NoError(p.T(), fixtures.Load())
	}
}

func (p *PostgresDatabaseTestSuite) getDSN(dbName string) string {
	var port nat.Port
	var err error

	if dbName == "walletstest" {
		port, err = p.walletsDBContainer.MappedPort(context.Background(), "5432")
	} else {
		port, err = p.transactionDBContainer.MappedPort(context.Background(), "5432")
	}

	p.Require().NoError(err)

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "maratest", "maratest", fmt.Sprintf("localhost:%s", port.Port()), dbName)
}

func (p *PostgresDatabaseTestSuite) TearDownTest() {
	for _, v := range databaseNames {

		db, err := sql.Open("postgres", p.getDSN(v))
		require.NoError(p.T(), err)

		err = db.Ping()
		require.NoError(p.T(), err)

		// drop all tables. not using migrate here as this seems to be faster
		// and will make these tests run faster
		_, err = db.Exec(`
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
	`)

		require.NoError(p.T(), err)
	}
}

func (p *PostgresDatabaseTestSuite) TestGetSenderWallet() {
	cfg := config.Configuration{
		Environment:             "local",
		WalletsPostgresDSN:      p.getDSN("walletstest"),
		TransactionsPostgresDSN: p.getDSN("transactionstest"),
	}

	db, err := New(cfg)
	require.NoError(p.T(), err)

	tt := []struct {
		valid   bool
		rowID   int64
		address string
	}{
		{
			valid:   true,
			rowID:   1,
			address: "mara.eth",
		},
		{
			valid:   true,
			rowID:   2,
			address: "chain.mara.eth",
		},
		{
			valid: false,
			rowID: 5,
		},
	}

	for _, v := range tt {

		wallet, err := db.GetWallet(context.Background(), v.rowID)
		if !v.valid {
			require.Error(p.T(), err)
			continue
		}

		require.NoError(p.T(), err)
		require.Equal(p.T(), v.address, strings.TrimRight(wallet.Address, " "))
	}
}

func (p *PostgresDatabaseTestSuite) TestGetDerivationPath() {
	cfg := config.Configuration{
		Environment:             "local",
		WalletsPostgresDSN:      p.getDSN("walletstest"),
		TransactionsPostgresDSN: p.getDSN("transactionstest"),
	}

	db, err := New(cfg)
	require.NoError(p.T(), err)

	// see testdata/fixtures/derivation_paths.yml
	tt := []struct {
		valid    bool
		wallet   string
		coinType uint32
		Account  uint32
		keyID    string
	}{
		{
			valid:    true,
			wallet:   "mara.eth",
			coinType: 614,
			Account:  4,
			keyID:    "MizBEqdhZ160syGCPFxYo6Lkxbiw",
		},
		{
			valid:    true,
			wallet:   "chain.mara.eth",
			coinType: 700,
			Account:  5,
			keyID:    "MizBEqdhZ160syGCPFfjfkjf",
		},
	}

	for _, v := range tt {
		derivationPath, err := db.GetDerivationPath(context.Background(), models.FindDerivationPathOptions{
			KeyID: v.keyID,
		})

		if !v.valid {
			require.Error(p.T(), err)
			continue
		}

		require.NoError(p.T(), err)
		require.Equal(p.T(), v.coinType, derivationPath.CoinType)
		require.Equal(p.T(), v.Account, derivationPath.Account)
	}
}

func (p *PostgresDatabaseTestSuite) TestCreateTransaction() {
	cfg := config.Configuration{
		WalletsPostgresDSN:      p.getDSN("walletstest"),
		TransactionsPostgresDSN: p.getDSN("transactionstest"),
		Environment:             "local",
	}

	db, err := New(cfg)
	require.NoError(p.T(), err)

	tx := &models.Transaction{
		SenderAddress:    "mara.eth",
		RecipientAddress: "mara.eth.link",
		State:            models.StateCreated,
		Nonce:            5,
		TRXHash:          "x984897550850750",
		NetworkType:      models.NetworkTypeMainnet,
		TransferType:     models.TransferTypeEoa,
		Amount:           big.NewFloat(100),
		ChainID:          10,
	}

	require.NoError(p.T(), db.CreateTransaction(context.Background(), tx))
}
