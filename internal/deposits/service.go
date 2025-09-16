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
		// Filtros defensivos: evita ensuciar DB si el parser aún no es real
		if e.Depositor == "" || len(e.Depositor) != 56 || e.Depositor[0] != 'G' {
			continue
		}
		if e.AmountRaw == "" || e.AmountRaw == "0" {
			continue
		}

		_, upsertErr := s.repo.Upsert(ctx, FunderDeposit{
			ContractID:     e.ContractID,
			Depositor:      e.Depositor,
			AmountRaw:      e.AmountRaw,
			OccurredAt:     time.Unix(e.OccurredAtUnix, 0).UTC(),
			ExternalID:     e.ExternalID, // txHash#opIndex
			TxHash:         e.TxHash,
			LedgerSequence: e.LedgerSequence,
			OpIndex:        e.OpIndex,
			Metadata:       e.Metadata,
		})
		if upsertErr != nil {
			// Importante: no abortar la corrida completa por un fallo puntual
			// (Puedes loguearlo aquí si ya tienes logger inyectado)
			continue
		}
	}

	// Devuelve el top N para ver el resultado actualizado
	return s.repo.ListByContract(ctx, contractID, 50)
}

func (s *Service) Repository() Repository { return s.repo }
