package deposits

import "time"

type FunderDeposit struct {
	ContractID     string         `json:"contractId"`
	Depositor      string         `json:"depositor"`
	AmountRaw      string         `json:"amount_raw"` // NUMERIC(39,0) como string
	OccurredAt     time.Time      `json:"occurred_at"`
	ExternalID     string         `json:"external_id,omitempty"` // txHash#opIndex
	TxHash         string         `json:"tx_hash,omitempty"`
	LedgerSequence int64          `json:"ledger_sequence,omitempty"`
	OpIndex        int32          `json:"op_index,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}
