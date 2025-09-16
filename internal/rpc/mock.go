package rpc

import (
	"context"
	"time"
)

type mockClient struct{}

func NewMockClient() Client { return &mockClient{} }

func (m *mockClient) FetchDeposits(ctx context.Context, contractID string) ([]DepositEvent, error) {
	now := time.Now().UTC()
	return []DepositEvent{
		{
			ContractID:     contractID,
			Depositor:      "GDUMMYDEPOSITOR111",
			AmountRaw:      "305000000",
			OccurredAtUnix: now.Add(-3 * time.Minute).Unix(),
			ExternalID:     "txhashA#0",
			TxHash:         "txhashA",
			LedgerSequence: 123456,
			OpIndex:        0,
			Metadata:       map[string]any{"source": "mock"},
		},
		{
			ContractID:     contractID,
			Depositor:      "GDUMMYDEPOSITOR222",
			AmountRaw:      "75000000",
			OccurredAtUnix: now.Add(-1 * time.Minute).Unix(),
			ExternalID:     "txhashB#1",
			TxHash:         "txhashB",
			LedgerSequence: 123457,
			OpIndex:        1,
			Metadata:       map[string]any{"source": "mock"},
		},
	}, nil
}
