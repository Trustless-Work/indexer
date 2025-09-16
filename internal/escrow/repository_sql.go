package escrow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type singleSQLRepo struct{ db *pgxpool.Pool }
type multiSQLRepo struct{ db *pgxpool.Pool }

func NewSingleSQLRepository(db *pgxpool.Pool) SingleRepository { return &singleSQLRepo{db} }
func NewMultiSQLRepository(db *pgxpool.Pool) MultiRepository   { return &multiSQLRepo{db} }

// SINGLE RELEASE ESCROW REPOSITORY
func (r *singleSQLRepo) CreateOrUpdate(ctx context.Context, e SingleReleaseJSON) error {
	if e.ContractID == "" {
		return errors.New("contractId requerido (PK)")
	}

	// Check if exists for update vs insert
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM single_release_escrow WHERE contract_id = $1)`
	err := r.db.QueryRow(ctx, checkQuery, e.ContractID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update using stored procedure (first delete, then insert)
		_, err = r.db.Exec(ctx, `SELECT sp_delete_full_single_release_escrow($1)`, e.ContractID)
		if err != nil {
			return err
		}
	}

	// Prepare data for stored procedure
	now := time.Now()
	
	// Convert milestones to JSONB
	milestonesJSON, err := json.Marshal([]map[string]interface{}{
		{
			"milestone_index": 0,
			"description":     e.Milestones[0].Description,
			"status":          "pending",
			"approved_at":     nil,
			"created_at":      now,
			"updated_at":      now,
		},
	})
	if err != nil {
		return err
	}

	// Call stored procedure for single release
	_, err = r.db.Exec(ctx, `
		SELECT insert_single_release_escrow_full(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21, 
			$22, $23, $24, $25, $26, $27, $28, $29, $30, $31, 
			$32, $33, $34, $35, $36
		)`,
		// Basic escrow data
		e.ContractID,          // p_contract_id
		e.ContractBaseID,      // p_contract_base_id  
		e.Description,         // p_description
		e.EngagementID,        // p_engagement_id
		e.PlatformFee,         // p_platform_fee
		e.ReceiverMemo,        // p_receiver_memo
		e.Signer,              // p_signer
		e.AmountRaw,           // p_amount  
		e.BalanceRaw,          // p_balance
		now,                   // p_created_at
		now,                   // p_updated_at
		
		// Trustline data
		e.Trustline.Address,   // p_trust_address
		e.Trustline.Decimals,  // p_trust_decimals
		e.Trustline.Name,      // p_trust_name
		now,                   // p_trust_created_at
		now,                   // p_trust_updated_at
		
		// Roles data
		e.Roles.Approver,         // p_approver
		e.Roles.DisputeResolver,  // p_dispute_resolver
		e.Roles.PlatformAddress,  // p_platform_address
		e.Roles.Receiver,         // p_receiver
		e.Roles.ReleaseSigner,    // p_release_signer
		e.Roles.ServiceProvider,  // p_service_provider
		now,                      // p_roles_created_at
		now,                      // p_roles_updated_at
		
		// Status data
		true,       // p_is_actived
		"active",   // p_status
		"",         // p_reason
		now,        // p_status_created_at
		now,        // p_status_updated_at
		
		// Flags data
		false,      // p_disputed
		false,      // p_released
		false,      // p_resolved
		now,        // p_flags_created_at
		now,        // p_flags_updated_at
		
		// Milestones and configs
		string(milestonesJSON), // p_milestones
		nil,                    // p_configs
	)
	
	return err
}

func (r *singleSQLRepo) Get(ctx context.Context, contractID string) (map[string]any, error) {
	// Use stored procedure to get complete data
	var result map[string]any
	err := r.db.QueryRow(ctx, `SELECT sp_select_single_release($1)`, contractID).Scan(&result)
	if err != nil {
		return nil, err
	}
	
	// Check if found
	if found, ok := result["found"].(bool); !ok || !found {
		return nil, errors.New("escrow not found")
	}
	
	return result, nil
}

func (r *singleSQLRepo) Delete(ctx context.Context, contractID string) error {
	// Use stored procedure for complete deletion
	var result map[string]any
	err := r.db.QueryRow(ctx, `SELECT sp_delete_full_single_release_escrow($1)`, contractID).Scan(&result)
	if err != nil {
		return err
	}
	
	// Check if existed
	if existed, ok := result["existed"].(bool); !ok || !existed {
		return errors.New("escrow not found")
	}
	
	return nil
}

// MULTI RELEASE ESCROW REPOSITORY  
func (r *multiSQLRepo) CreateOrUpdate(ctx context.Context, e MultiReleaseJSON) error {
	if e.ContractID == "" {
		return errors.New("contractId requerido (PK)")
	}

	// Check if exists for update vs insert
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM multi_release_escrow WHERE contract_id = $1)`
	err := r.db.QueryRow(ctx, checkQuery, e.ContractID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update using stored procedure (first delete, then insert)
		_, err = r.db.Exec(ctx, `SELECT sp_delete_full_multi_release_escrow($1)`, e.ContractID)
		if err != nil {
			return err
		}
	}

	// Prepare data for stored procedure
	now := time.Now()
	
	// Convert milestones to JSONB
	milestonesData := make([]map[string]interface{}, len(e.Milestones))
	for i, milestone := range e.Milestones {
		milestonesData[i] = map[string]interface{}{
			"milestone_index": i,
			"description":     milestone.Description,
			"amount":          milestone.Amount,
			"balance":         milestone.Amount, // Initially balance equals amount
			"status":          "pending",
			"approved_at":     nil,
			"created_at":      now,
			"updated_at":      now,
		}
	}
	
	milestonesJSON, err := json.Marshal(milestonesData)
	if err != nil {
		return err
	}

	// Call stored procedure for multi release
	_, err = r.db.Exec(ctx, `
		SELECT insert_multi_release_escrow_full(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, 
			$10, $11, $12, $13, $14, $15, $16, $17, $18, $19, 
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		)`,
		// Basic escrow data
		e.ContractID,          // p_contract_id
		e.ContractBaseID,      // p_contract_base_id
		e.Description,         // p_description
		e.EngagementID,        // p_engagement_id
		e.PlatformFee,         // p_platform_fee
		e.ReceiverMemo,        // p_receiver_memo
		e.Signer,              // p_signer
		now,                   // p_created_at
		now,                   // p_updated_at
		
		// Trustline data  
		e.Trustline.Address,   // p_trust_address
		e.Trustline.Decimals,  // p_trust_decimals
		e.Trustline.Name,      // p_trust_name
		now,                   // p_trust_created_at
		now,                   // p_trust_updated_at
		
		// Roles data
		e.Roles.Approver,         // p_approver
		e.Roles.DisputeResolver,  // p_dispute_resolver
		e.Roles.PlatformAddress,  // p_platform_address
		e.Roles.Receiver,         // p_receiver
		e.Roles.ReleaseSigner,    // p_release_signer
		e.Roles.ServiceProvider,  // p_service_provider
		now,                      // p_roles_created_at
		now,                      // p_roles_updated_at
		
		// Status data
		true,       // p_is_actived
		"active",   // p_status
		"",         // p_reason
		now,        // p_status_created_at
		now,        // p_status_updated_at
		
		// Milestones and configs
		string(milestonesJSON), // p_milestones
		nil,                    // p_configs
	)
	
	return err
}

func (r *multiSQLRepo) Get(ctx context.Context, contractID string) (map[string]any, error) {
	// Use stored procedure to get complete data
	var result map[string]any
	err := r.db.QueryRow(ctx, `SELECT sp_select_multi_release($1)`, contractID).Scan(&result)
	if err != nil {
		return nil, err
	}
	
	// Check if found
	if found, ok := result["found"].(bool); !ok || !found {
		return nil, errors.New("escrow not found")
	}
	
	return result, nil
}

func (r *multiSQLRepo) Delete(ctx context.Context, contractID string) error {
	// Use stored procedure for complete deletion
	var result map[string]any
	err := r.db.QueryRow(ctx, `SELECT sp_delete_full_multi_release_escrow($1)`, contractID).Scan(&result)
	if err != nil {
		return err
	}
	
	// Check if existed
	if existed, ok := result["existed"].(bool); !ok || !existed {
		return errors.New("escrow not found")
	}
	
	return nil
}