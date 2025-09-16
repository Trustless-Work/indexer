package deposits

import "context"

type Repository interface {
	Upsert(ctx context.Context, d FunderDeposit) (string, error)
	ListByContract(ctx context.Context, contractID string, limit int) ([]FunderDeposit, error)
}
