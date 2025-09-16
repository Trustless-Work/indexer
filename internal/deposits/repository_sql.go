package deposits

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlRepo struct{ db *pgxpool.Pool }

func NewSQLRepository(db *pgxpool.Pool) Repository { return &sqlRepo{db} }

// Upsert llama al SP idempotente y retorna la acción ("inserted|ignored|updated").
func (r *sqlRepo) Upsert(ctx context.Context, d FunderDeposit) (string, error) {
	const q = `
SELECT sp_insert_funder_deposit(
	$1::text,        -- contract_id
	$2::text,        -- depositor
	$3::numeric,     -- amount_raw
	$4::timestamptz, -- occurred_at
	$5::text,        -- external_id (txHash#opIndex)
	$6::text,        -- tx_hash
	$7::bigint,      -- ledger_sequence
	$8::int,         -- op_index
	$9::jsonb        -- metadata
)::jsonb;
`
	var raw []byte
	if err := r.db.QueryRow(ctx, q,
		d.ContractID,
		d.Depositor,
		d.AmountRaw,
		d.OccurredAt,
		d.ExternalID,
		d.TxHash,
		d.LedgerSequence,
		d.OpIndex,
		d.Metadata,
	).Scan(&raw); err != nil {
		return "", err
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}

	action, _ := out["action"].(string)
	if action == "" {
		return "", errors.New("sp_insert_funder_deposit: missing action")
	}
	return action, nil
}

func (r *sqlRepo) ListByContract(ctx context.Context, contractID string, limit int) ([]FunderDeposit, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
SELECT contract_id,
       depositor,
       amount_raw::text,
       occurred_at,
       external_id,
       tx_hash,
       ledger_sequence,
       op_index,
       metadata
FROM escrow_funder_deposits
WHERE contract_id = $1
ORDER BY occurred_at DESC
LIMIT $2;
`
	rows, err := r.db.Query(ctx, q, contractID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]FunderDeposit, 0, limit)
	for rows.Next() {
		var (
			d   FunderDeposit
			occ time.Time
			js  []byte
		)
		if err := rows.Scan(
			&d.ContractID,
			&d.Depositor,
			&d.AmountRaw,
			&occ,
			&d.ExternalID,
			&d.TxHash,
			&d.LedgerSequence,
			&d.OpIndex,
			&js,
		); err != nil {
			return nil, err
		}
		d.OccurredAt = occ
		if len(js) > 0 {
			var md map[string]any
			_ = json.Unmarshal(js, &md) // tolerante
			d.Metadata = md
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
