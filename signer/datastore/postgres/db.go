// Package postgres implements a storage backend backed by postgres
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ericlagergren/decimal"
	"github.com/lib/pq"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"

	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/datastore/postgres/models/transactions"
	dbmodels "github.com/mara-labs/transactionsigner/datastore/postgres/models/wallets"
	"github.com/mara-labs/transactionsigner/models"
)

// Store is a wrapper for sqlboiler which encapsulates all DB operations
type Store struct {
	walletsDB      *sql.DB
	transactionsDB *sql.DB
}

// New creates a new store backed by Postgres
func New(cfg config.Configuration) (*Store, error) {
	sqltrace.Register("pq", &pq.Driver{})

	db, err := sqltrace.Open("pq", cfg.TransactionsPostgresDSN)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping transactions db...%v", err)
	}

	walletsDB, err := sqltrace.Open("pq", cfg.WalletsPostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("could not ping wallets db...%v", err)
	}

	if err := walletsDB.Ping(); err != nil {
		return nil, err
	}

	boil.DebugMode = config.IsLocal(cfg)

	return &Store{
		transactionsDB: db,
		walletsDB:      walletsDB,
	}, nil
}

// Close shuts down the underlying db connections
func (s *Store) Close() error {
	if err := s.walletsDB.Close(); err != nil {
		return err
	}

	return s.transactionsDB.Close()
}

// GetDerivationPath retrieves the bip32 path for the given keyID
func (s *Store) GetDerivationPath(ctx context.Context, opts models.FindDerivationPathOptions) (*models.DerivationPath, error) {
	derivedPathFromDB, err := dbmodels.DerivationPaths(
		qm.InnerJoin(`sender_wallets as w ON w.id = "derivation_paths".wallet_id AND w.key_id = ?`, opts.KeyID)).
		One(ctx, s.walletsDB)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDerivationPathNotFound
	}

	if err != nil {
		return nil, err
	}

	return &models.DerivationPath{
		ID:       int64(derivedPathFromDB.ID),
		WalletID: uint32(derivedPathFromDB.WalletID.Int),
		Purpose:  uint32(derivedPathFromDB.Purpose),
		CoinType: uint32(derivedPathFromDB.CoinType),
		Account:  uint32(derivedPathFromDB.Account),
		Change:   uint32(derivedPathFromDB.Change),
		Created:  derivedPathFromDB.Created,
		Updated:  derivedPathFromDB.Updated,
	}, nil
}

// GetWallet retrieves the wallet that has the accompanying address
// the wallet address is canonicalized to lower case before it is searched
func (s *Store) GetWallet(ctx context.Context, walletRowID int64) (*models.SenderWallet, error) {
	retrievedWallet, err := dbmodels.SenderWallets(qm.Where("id = ?", walletRowID)).
		One(ctx, s.walletsDB)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrWalletNotFound
	}

	if err != nil {
		return nil, err
	}

	return &models.SenderWallet{
		UserID:       retrievedWallet.UserID.Int,
		ID:           int64(retrievedWallet.ID),
		Address:      retrievedWallet.Address,
		ChainID:      retrievedWallet.ChainID,
		KeyID:        retrievedWallet.KeyID,
		IsMultisig:   retrievedWallet.IsMultisig,
		Created:      retrievedWallet.Created,
		Updated:      retrievedWallet.Updated,
		AddressIndex: retrievedWallet.AddressIndex,
	}, nil
}

// CreateTransaction persists a transaction into the db
func (s *Store) CreateTransaction(ctx context.Context, trans *models.Transaction) error {
	amount := decimal.New(0, 0).SetFloat(trans.Amount)

	t := &transactions.Transaction{
		ID:                trans.ID,
		ChainID:           int(trans.ChainID),
		TRXHash:           trans.TRXHash,
		SenderAddress:     trans.SenderAddress,
		RecipientAddress:  trans.RecipientAddress,
		Nonce:             int(trans.Nonce),
		IsSenderPayingGas: trans.IsSenderPayingGas,
		Created:           trans.Created,
		Updated:           trans.Updated,
		NetworkType:       transactions.NetworkType(trans.NetworkType),
		State:             transactions.State(trans.State),
		TransferType:      transactions.TransferType(trans.TransferType),
		MaxFee:            types.NewNullDecimal(decimal.New(trans.MaxFee, 0)),
		MaxPriorityFee:    types.NewNullDecimal(decimal.New(trans.MaxPriorityFee, 0)),
		Amount:            types.NewDecimal(amount),
	}

	return t.Insert(ctx, s.transactionsDB, boil.Infer())
}
