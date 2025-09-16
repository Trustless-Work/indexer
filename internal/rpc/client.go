package rpc

import "context"

type DepositEvent struct {
	ContractID     string
	Depositor      string
	AmountRaw      string
	OccurredAtUnix int64
	ExternalID     string
	TxHash         string
	LedgerSequence int64
	OpIndex        int32
	Metadata       map[string]any
}

type Client interface {
	FetchDeposits(ctx context.Context, contractID string) ([]DepositEvent, error)
}

func NewHTTPClient(baseURL string) Client { 
	return NewSimpleStellarRPCClient(baseURL)
}

type httpClient struct{ baseURL string }

// TODO: implementar getEvents real
func (c *httpClient) FetchDeposits(ctx context.Context, contractID string) ([]DepositEvent, error) {
	return nil, nil
}
