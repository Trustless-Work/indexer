package deposits

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlRepo struct{ db *pgxpool.Pool }

func NewSQLRepository(db *pgxpool.Pool) Repository { return &sqlRepo{db} }

func (r *sqlRepo) Upsert(ctx context.Context, d FunderDeposit) (string, error) {
	const q = `
INSERT INTO escrow_funder_deposits
  (contract_id, depositor, amount_raw, occurred_at, external_id, tx_hash, ledger_sequence, op_index, metadata)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (contract_id, external_id) DO UPDATE SET
  depositor=EXCLUDED.depositor,
  amount_raw=EXCLUDED.amount_raw,
  occurred_at=EXCLUDED.occurred_at,
  tx_hash=EXCLUDED.tx_hash,
  ledger_sequence=EXCLUDED.ledger_sequence,
  op_index=EXCLUDED.op_index,
  metadata=EXCLUDED.metadata
`
	_, err := r.db.Exec(ctx, q,
		d.ContractID, d.Depositor, d.AmountRaw, d.OccurredAt,
		d.ExternalID, d.TxHash, d.LedgerSequence, d.OpIndex, d.Metadata,
	)
	return "upserted", err
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
