package escrow

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type singleSQLRepo struct{ db *pgxpool.Pool }
type multiSQLRepo struct{ db *pgxpool.Pool }

func NewSingleSQLRepository(db *pgxpool.Pool) SingleRepository { return &singleSQLRepo{db} }
func NewMultiSQLRepository(db *pgxpool.Pool) MultiRepository   { return &multiSQLRepo{db} }

// Ajusta nombres de columnas/tabla a tu DDL real.
// Guardamos roles/milestones/trustline como JSONB por simplicidad inicial.
func (r *singleSQLRepo) CreateOrUpdate(ctx context.Context, e SingleReleaseJSON) error {
	if e.ContractID == "" {
		return errors.New("contractId requerido (PK)")
	}
	const q = `
INSERT INTO single_release_escrows
  (contract_id, signer, engagement_id, title, description, roles, amount, platform_fee, milestones, trustline, receiver_memo)
VALUES ($1,$2,$3,$4,$5, to_jsonb($6::json), $7,$8, to_jsonb($9::json), to_jsonb($10::json), $11)
ON CONFLICT (contract_id) DO UPDATE SET
  signer=EXCLUDED.signer,
  engagement_id=EXCLUDED.engagement_id,
  title=EXCLUDED.title,
  description=EXCLUDED.description,
  roles=EXCLUDED.roles,
  amount=EXCLUDED.amount,
  platform_fee=EXCLUDED.platform_fee,
  milestones=EXCLUDED.milestones,
  trustline=EXCLUDED.trustline,
  receiver_memo=EXCLUDED.receiver_memo
`
	roles := map[string]any{
		"approver": e.Roles.Approver, "serviceProvider": e.Roles.ServiceProvider,
		"platformAddress": e.Roles.PlatformAddress, "releaseSigner": e.Roles.ReleaseSigner,
		"disputeResolver": e.Roles.DisputeResolver, "receiver": e.Roles.Receiver,
	}
	_, err := r.db.Exec(ctx, q,
		e.ContractID, e.Signer, e.EngagementID, e.Title, e.Description,
		roles, e.Amount, e.PlatformFee, e.Milestones, e.Trustline, e.ReceiverMemo,
	)
	return err
}

func (r *singleSQLRepo) Get(ctx context.Context, contractID string) (map[string]any, error) {
	const q = `SELECT to_jsonb(t) FROM single_release_escrows t WHERE contract_id=$1`
	var j map[string]any
	if err := r.db.QueryRow(ctx, q, contractID).Scan(&j); err != nil {
		return nil, err
	}
	return j, nil
}

func (r *singleSQLRepo) Delete(ctx context.Context, contractID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM single_release_escrows WHERE contract_id=$1`, contractID)
	return err
}

// MULTI
func (r *multiSQLRepo) CreateOrUpdate(ctx context.Context, e MultiReleaseJSON) error {
	if e.ContractID == "" {
		return errors.New("contractId requerido (PK)")
	}
	const q = `
INSERT INTO multi_release_escrows
  (contract_id, signer, engagement_id, title, description, roles, platform_fee, milestones, trustline, receiver_memo)
VALUES ($1,$2,$3,$4,$5, to_jsonb($6::json), $7, to_jsonb($8::json), to_jsonb($9::json), $10)
ON CONFLICT (contract_id) DO UPDATE SET
  signer=EXCLUDED.signer,
  engagement_id=EXCLUDED.engagement_id,
  title=EXCLUDED.title,
  description=EXCLUDED.description,
  roles=EXCLUDED.roles,
  platform_fee=EXCLUDED.platform_fee,
  milestones=EXCLUDED.milestones,
  trustline=EXCLUDED.trustline,
  receiver_memo=EXCLUDED.receiver_memo
`
	roles := map[string]any{
		"approver": e.Roles.Approver, "serviceProvider": e.Roles.ServiceProvider,
		"platformAddress": e.Roles.PlatformAddress, "releaseSigner": e.Roles.ReleaseSigner,
		"disputeResolver": e.Roles.DisputeResolver, "receiver": e.Roles.Receiver,
	}
	_, err := r.db.Exec(ctx, q,
		e.ContractID, e.Signer, e.EngagementID, e.Title, e.Description,
		roles, e.PlatformFee, e.Milestones, e.Trustline, e.ReceiverMemo,
	)
	return err
}

func (r *multiSQLRepo) Get(ctx context.Context, contractID string) (map[string]any, error) {
	const q = `SELECT to_jsonb(t) FROM multi_release_escrows t WHERE contract_id=$1`
	var j map[string]any
	if err := r.db.QueryRow(ctx, q, contractID).Scan(&j); err != nil {
		return nil, err
	}
	return j, nil
}

func (r *multiSQLRepo) Delete(ctx context.Context, contractID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM multi_release_escrows WHERE contract_id=$1`, contractID)
	return err
}
