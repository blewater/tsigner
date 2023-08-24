// Package models defines models that are used
package models

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

// SignOptions defines a set of properties that can be used to retrieve the right signing details
type SignOptions struct {
	KeyID          string
	TX             *types.Transaction
	DerivationPath []uint32
}

// Signer models a way to sign a given transaction. Right now we only provide a Sepior implementation
type Signer interface {
	Sign(context.Context, SignOptions) (*types.Transaction, error)
}
