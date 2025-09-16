package deposits

import (
	"context"
	"time"

	"github.com/Trustless-Work/indexer/internal/rpc"
)

type Service struct {
	repo Repository
	rpc  rpc.Client
}

func NewService(repo Repository, r rpc.Client) *Service {
	return &Service{repo: repo, rpc: r}
}

func (s *Service) IndexContractDeposits(ctx context.Context, contractID string) ([]FunderDeposit, error) {
	events, err := s.rpc.FetchDeposits(ctx, contractID)
	if err != nil {
		return nil, err
	}

	for _, e := range events {
		_, err := s.repo.Upsert(ctx, FunderDeposit{
			ContractID:     e.ContractID,
			Depositor:      e.Depositor,
			AmountRaw:      e.AmountRaw,
			OccurredAt:     time.Unix(e.OccurredAtUnix, 0).UTC(),
			ExternalID:     e.ExternalID,
			TxHash:         e.TxHash,
			LedgerSequence: e.LedgerSequence,
			OpIndex:        e.OpIndex,
			Metadata:       e.Metadata,
		})
		if err != nil {
			return nil, err
		}
	}
	return s.repo.ListByContract(ctx, contractID, 50)
}
