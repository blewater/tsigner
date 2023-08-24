package models

import (
	"context"
	"io"
)

// CreatedTxQueueItem models the data structure received from the queue
type CreatedTxQueueItem struct {
	ID            string `json:"id,omitempty"`
	RawTX         string `json:"raw_tx,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	WalletRowID   int64  `json:"wallet_row_id,omitempty"`
}

// SignedTXQueueItem models the data structure for transactions to be broadcasted onchain
type SignedTXQueueItem struct {
	ID            string `json:"id"`
	SignedTX      string `json:"signed_tx"`
	TransactionID string `json:"transaction_id,omitempty"`
}

// Queue implements a set of methods to add new items to a queue
type Queue interface {
	io.Closer
	Add(context.Context, SignedTXQueueItem) error
}
