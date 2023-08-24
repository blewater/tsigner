package models

import (
	"context"
	"errors"
	"io"
	"math/big"
	"time"
)

var (
	// ErrWalletNotFound is a custom error that can be used when a wallet doesn't exists in the datastore
	ErrWalletNotFound = errors.New("wallet not found")
	// ErrDerivationPathNotFound is a custom error that can be used instead of
	// datastore specific errors
	ErrDerivationPathNotFound = errors.New("derivation path not found")
)

const (
	// AddressTypeEoa is an address that is controlled by an external party
	// with a secret/phrase
	AddressTypeEoa AddressType = "eoa"
	// AddressTypeSmartContract is an address that is essentially a smart
	// contract
	AddressTypeSmartContract AddressType = "smart_contract"
	// AddressTypeMultisig is an address that is a multisig
	AddressTypeMultisig AddressType = "multisig"
)

const (
	// NetworkTypeMainnet is mainnet
	NetworkTypeMainnet NetworkType = "mainnet"
	// NetworkTypeTestnet is a testnet
	NetworkTypeTestnet NetworkType = "testnet"
)

// State is used to identify the current status of a transaction in the cycle
type State string

// Enum values for State
const (
	StateInitiated State = "initiated"
	StateCreated   State = "created"
	StateSigned    State = "signed"
	StateSubmitted State = "submitted"
	StateSettled   State = "settled"
	StateFinalized State = "finalized"
	StateErred     State = "erred"
)

// TransferType denotes the transaction type
type TransferType string

// Enum values for TransferType
const (
	TransferTypeEoa           TransferType = "eoa"
	TransferTypeSmartContract TransferType = "smart_contract"
)

// NetworkType denotes the network we are on
type NetworkType string

// AddressType is used to identify the type of address currently in use
type AddressType string

// DerivationPath models the derivation_path table
type DerivationPath struct {
	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`

	ID           int64  `json:"id,omitempty"`
	WalletID     uint32 `json:"wallet_id,omitempty"`
	Purpose      uint32 `json:"purpose,omitempty"`
	CoinType     uint32 `json:"coin_type,omitempty"`
	Account      uint32 `json:"account,omitempty"`
	Change       uint32 `json:"change,omitempty"`
	AddressIndex uint32 `json:"address_index,omitempty"`
}

// SenderWallet is an object representing the wallets table.
type SenderWallet struct {
	ID                int64     `json:"id,omitempty"`
	UserID            int       `json:"user_id,omitempty"`
	Address           string    `json:"address,omitempty"`
	ChainID           int       `json:"chain_id,omitempty"`
	KeyID             string    `json:"key_id,omitempty"`
	IsMultisig        bool      `json:"is_multisig,omitempty"`
	AddressIndex      int       `json:"address_index,omitempty"`
	MultisigThreshold int64     `json:"multisig_threshold,omitempty"`
	Created           time.Time `json:"created,omitempty"`
	Updated           time.Time `json:"updated,omitempty"`
}

// FindDerivationPathOptions defines properties that can be used to retrieve a bip32 path
type FindDerivationPathOptions struct {
	KeyID string
}

// Transaction is an object representing the tx table
type Transaction struct {
	ID                int          `json:"id"`
	TRXHash           string       `json:"trx_hash"`
	ChainID           int64        `json:"chain_id"`
	NetworkType       NetworkType  `json:"network_type"`
	State             State        `json:"state"`
	TransferType      TransferType `json:"transfer_type"`
	SenderAddress     string       `json:"sender_address"`
	RecipientAddress  string       `json:"recipient_address"`
	Amount            *big.Float   `json:"amount"`
	Nonce             int64        `json:"nonce"`
	MaxFee            int64        `json:"max_fee"`
	MaxPriorityFee    int64        `json:"max_priority_fee"`
	IsSenderPayingGas bool         `json:"is_sender_paying_gas"`
	Created           time.Time    `json:"created"`
	Updated           time.Time    `json:"updated"`
}

// Datastore is an interface for persisting and retriveing data
type Datastore interface {
	io.Closer
	GetDerivationPath(context.Context, FindDerivationPathOptions) (*DerivationPath, error)
	GetWallet(context.Context, int64) (*SenderWallet, error)
	CreateTransaction(context.Context, *Transaction) error
}
