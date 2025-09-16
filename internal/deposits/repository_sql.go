package deposits

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlRepo struct{ db *pgxpool.Pool }

func NewSQLRepository(db *pgxpool.Pool) Repository { return &sqlRepo{db} }

func (r *sqlRepo) Upsert(ctx context.Context, d FunderDeposit) (string, error) {
	// Use stored procedure for better validation and consistency
	var result map[string]any
	err := r.db.QueryRow(ctx, `
		SELECT sp_insert_funder_deposit($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		d.ContractID, d.Depositor, d.AmountRaw, d.OccurredAt,
		d.ExternalID, d.TxHash, d.LedgerSequence, d.OpIndex, d.Metadata,
	).Scan(&result)
	
	if err != nil {
		return "", err
	}
	
	// Extract insert_id from result if available
	if insertID, ok := result["insert_id"]; ok {
		return insertID.(string), nil
	}
	
	return "upserted", nil
}

func (r *sqlRepo) ListByContract(ctx context.Context, contractID string, limit int) ([]FunderDeposit, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
SELECT contract_id, depositor, amount_raw::text, occurred_at, external_id, tx_hash, ledger_sequence, op_index, metadata
FROM escrow_funder_deposits
WHERE contract_id=$1
ORDER BY occurred_at DESC
LIMIT $2
`
	rows, err := r.db.Query(ctx, q, contractID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []FunderDeposit{}
	for rows.Next() {
		var d FunderDeposit
		var md map[string]any
		var occ time.Time
		if err := rows.Scan(&d.ContractID, &d.Depositor, &d.AmountRaw, &occ, &d.ExternalID, &d.TxHash, &d.LedgerSequence, &d.OpIndex, &md); err != nil {
			return nil, err
		}
		d.OccurredAt = occ
		d.Metadata = md
		out = append(out, d)
	}
	return out, rows.Err()
}
