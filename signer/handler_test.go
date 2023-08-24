package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/mara-labs/transactionsigner/config"
	"github.com/mara-labs/transactionsigner/mocks"
	"github.com/mara-labs/transactionsigner/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewHandler(t *testing.T) {
	f, err := os.Open("testdata/event.json")
	require.NoError(t, err)

	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()

	var eventData events.SQSEvent

	err = json.NewDecoder(f).Decode(&eventData)
	require.NoError(t, err)

	tt := []struct {
		name     string
		hasError bool
		event    events.SQSEvent
		mockFn   func(*mocks.Store, *mocks.MockQueue, *mocks.MockSigner)
	}{
		{
			name: "no error because lambda has zero items",
			mockFn: func(_ *mocks.Store, _ *mocks.MockQueue, _ *mocks.MockSigner) {
			},
		},
		{
			name: "could not get wallet",
			mockFn: func(s *mocks.Store, _ *mocks.MockQueue, _ *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(nil, errors.New("could not fetch wallet"))
			},
			hasError: true,
			event:    eventData,
		},
		{
			name: "wallet was retrieved but derivation path could not be retrieved",
			mockFn: func(s *mocks.Store, _ *mocks.MockQueue, _ *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(nil, errors.New("could not fetch derivation path"))
			},
			hasError: true,
			event:    eventData,
		},
		{
			name: "wallet was retrieved but derivation path could not be retrieved",
			mockFn: func(s *mocks.Store, _ *mocks.MockQueue, _ *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(nil, errors.New("could not fetch derivation path"))
			},
			hasError: true,
			event:    eventData,
		},
		{
			name: "could not sign tx",
			mockFn: func(s *mocks.Store, _ *mocks.MockQueue, signer *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.DerivationPath{
						Account: 614,
					}, nil)

				signer.EXPECT().Sign(gomock.Any(), gomock.Any()).Times(1).
					Return(nil, errors.New("failed to sign TX"))
			},
			hasError: true,
			event:    eventData,
		},
		{
			name: "signed TX but could not create tx",
			mockFn: func(s *mocks.Store, _ *mocks.MockQueue, signer *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.DerivationPath{
						Account: 614,
					}, nil)

				signer.EXPECT().Sign(gomock.Any(), gomock.Any()).Times(1).
					Return(types.
						NewTransaction(0, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(0), 0, big.NewInt(0), nil),
						nil)

				s.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Times(1).
					Return(errors.New("could not create TX"))
			},
			hasError: true,
			event:    eventData,
		},
		{
			name: "transaction was stored but could not be added to queue",
			mockFn: func(s *mocks.Store, queue *mocks.MockQueue, signer *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.DerivationPath{
						Account: 614,
					}, nil)

				signer.EXPECT().Sign(gomock.Any(), gomock.Any()).Times(1).
					Return(types.
						NewTransaction(0, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(0), 0, big.NewInt(0), nil),
						nil)

				s.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Times(1).
					Return(nil)

				queue.EXPECT().Add(gomock.Any(), gomock.Any()).Times(1).
					Return(errors.New("could not add item to the queue"))
			},
			hasError: true,
			event:    eventData,
		},
		{
			name: "TX was added to the queue but the db store could not be closed",
			mockFn: func(s *mocks.Store, queue *mocks.MockQueue, signer *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.DerivationPath{
						Account: 614,
					}, nil)

				signer.EXPECT().Sign(gomock.Any(), gomock.Any()).Times(1).
					Return(types.
						NewTransaction(0, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(0), 0, big.NewInt(0), nil),
						nil)

				s.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Times(1).
					Return(nil)

				queue.EXPECT().Add(gomock.Any(), gomock.Any()).Times(1).
					Return(nil)

				s.EXPECT().Close().Times(1).Return(nil)
			},
			event: eventData,
		},
		{
			name: "Handler runs successfully",
			mockFn: func(s *mocks.Store, queue *mocks.MockQueue, signer *mocks.MockSigner) {
				s.EXPECT().GetWallet(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.SenderWallet{
						Address: "mara.eth",
					}, nil)

				s.EXPECT().GetDerivationPath(gomock.Any(), gomock.Any()).Times(1).
					Return(&models.DerivationPath{
						Account: 614,
					}, nil)

				signer.EXPECT().Sign(gomock.Any(), gomock.Any()).Times(1).
					Return(types.
						NewTransaction(0, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(0), 0, big.NewInt(0), nil),
						nil)

				s.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Times(1).
					Return(nil)

				queue.EXPECT().Add(gomock.Any(), gomock.Any()).Times(1).
					Return(nil)

				s.EXPECT().Close().Times(1).Return(nil)
			},
			hasError: false,
			event:    eventData,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mocks.NewStore(ctrl)
			queue := mocks.NewMockQueue(ctrl)
			signer := mocks.NewMockSigner(ctrl)

			cfg := config.Configuration{
				LogLevel: "DEBUG",
			}

			v.mockFn(store, queue, signer)

			handler := newHandler(cfg, signer, queue, store)

			ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
				AwsRequestID: uuid.New().String(),
			})

			err := handler(ctx, v.event)

			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
