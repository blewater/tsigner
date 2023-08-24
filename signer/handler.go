// package main implements a lambda handler that signs transactions
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	chainutil "github.com/mara-labs/chain-util"

	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/models"
)

// LambdaHandler is a type that denotes a valid lambda function for our lambdas integration
type LambdaHandler func(context.Context, events.SQSEvent) error

func getItemFromQueueBody(record events.SQSMessage,
) (*models.CreatedTxQueueItem, error) {
	item := new(models.CreatedTxQueueItem)

	if err := json.NewDecoder(strings.NewReader(record.Body)).
		Decode(item); err != nil {
		log.WithField("message_id", record.MessageId).
			WithField("receipt_handle", record.ReceiptHandle).
			WithError(err).
			Error("could not decode message from the queue")

		return nil, err
	}

	return item, nil
}

func newHandler(configValues config.Configuration,
	signer models.Signer, queue models.Queue,
	datastore models.Datastore,
) LambdaHandler {
	return func(ctx context.Context, event events.SQSEvent) error {
		if len(event.Records) == 0 {
			return nil
		}

		record := event.Records[0]

		lambdaContext, ok := lambdacontext.FromContext(ctx)
		if !ok {
			return errors.New("context is invalid")
		}

		log.SetFormatter(&log.JSONFormatter{})

		item, err := getItemFromQueueBody(record)
		if err != nil {
			log.WithError(err).
				Error("could not decode queue item")
			return err
		}

		logger := log.WithField("request_id", lambdaContext.AwsRequestID).
			WithField("transaction_id", item.TransactionID)

		level, err := getLevel(configValues)
		if err != nil {
			logger.WithError(err).Error("could not load the right log level")
			return err
		}

		logger.Logger.SetLevel(level)

		tracer.Start(
			tracer.WithService(configValues.ServiceName),
			tracer.WithEnv(configValues.Environment),
		)

		span, spanCtx := tracer.StartSpanFromContext(ctx, "Transaction.SignTransaction")

		span.SetTag("request_id", lambdaContext.AwsRequestID)
		span.SetTag("transaction_id", item.TransactionID)
		span.SetTag("request_message_id", record.MessageId)
		span.SetTag("request_message_receipt_handle", record.ReceiptHandle)

		defer span.Finish()

		tx := &types.Transaction{}

		rawTxBytes, err := hex.DecodeString(item.RawTX)
		if err != nil {
			logger.WithField("message_id", record.MessageId).
				WithField("receipt_handle", record.ReceiptHandle).
				WithError(err).
				WithField("raw_tx", item.RawTX).
				Error("could not hex decode raw TX string")
			return err
		}

		err = rlp.DecodeBytes(rawTxBytes, tx)
		if err != nil {
			logger.WithField("message_id", record.MessageId).
				WithField("receipt_handle", record.ReceiptHandle).
				WithError(err).
				WithField("raw_tx", item.RawTX).
				Error("could not unmarshal raw tx bytes")
			return err
		}

		wallet, err := datastore.GetWallet(spanCtx, item.WalletRowID)
		if err != nil {
			logger.WithField("message_id", record.MessageId).
				WithField("receipt_handle", record.ReceiptHandle).
				WithError(err).
				WithField("raw_tx", item.RawTX).
				Error("could not fetch wallet address")
			return err
		}

		derivationPath, err := datastore.GetDerivationPath(spanCtx, models.FindDerivationPathOptions{
			KeyID: wallet.KeyID,
		})
		if err != nil {
			logger.WithField("message_id", record.MessageId).
				WithField("receipt_handle", record.ReceiptHandle).
				WithError(err).
				WithField("raw_tx", item.RawTX).
				WithField("key_id", wallet.KeyID).
				Error("could not fetch derivation path")
			return err
		}

		signedTX, err := signer.Sign(spanCtx, models.SignOptions{
			KeyID: wallet.KeyID,
			TX:    tx,
			DerivationPath: []uint32{
				derivationPath.Purpose, derivationPath.CoinType, derivationPath.Account,
				derivationPath.Change, uint32(wallet.AddressIndex),
			},
		})
		if err != nil {
			logger.WithError(err).
				WithField("key_id", wallet.KeyID).
				WithField("hash", tx.Hash()).
				WithField("transaction_id", item.TransactionID).
				Error("could not sign transaction")
			return err
		}

		ts := types.Transactions{signedTX}
		b := new(bytes.Buffer)
		ts.EncodeIndex(0, b)

		amount := tx.Value()

		amountToAdd, err := chainutil.BigIntToFloat64(amount)
		if err != nil {
			logger.WithField("amount", amount).
				WithError(err).
				WithField("message_id", amount).
				Error("could not parse amount")

			return err
		}

		dbTransaction := &models.Transaction{
			TRXHash:           signedTX.Hash().Hex(),
			ChainID:           signedTX.ChainId().Int64(),
			IsSenderPayingGas: false,
			SenderAddress:     wallet.Address,
			Nonce:             int64(signedTX.Nonce()),
			RecipientAddress:  signedTX.To().Hex(),
			State:             models.StateSigned,
			NetworkType:       models.NetworkTypeMainnet,
			TransferType:      models.TransferTypeEoa,
			Amount:            amountToAdd,
		}

		if err := datastore.CreateTransaction(ctx, dbTransaction); err != nil {
			logger.WithError(err).Error("could not create transaction")
			return err
		}

		err = queue.Add(spanCtx, models.SignedTXQueueItem{
			ID:            record.MessageId,
			SignedTX:      hex.EncodeToString(b.Bytes()),
			TransactionID: item.TransactionID,
		})

		if err != nil {
			logger.WithError(err).
				Error("could not add signed item to the queue")
			return err
		}

		if err := datastore.Close(); err != nil {
			logger.WithError(err).Error("could not close database connection")
			return nil
		}

		return nil
	}
}
