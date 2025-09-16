package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SimpleStellarRPCClient: cliente RPC simple (sin decodificar XDR aún)
type simpleStellarRPCClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSimpleStellarRPCClient(baseURL string) Client {
	return &simpleStellarRPCClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ----------- Tipos RPC genéricos -----------

type SimpleRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
}

type SimpleRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *SimpleRPCError `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type SimpleRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// ----------- getEvents -----------

type SimpleEventsRequest struct {
	StartLedger uint32                  `json:"startLedger"`
	EndLedger   uint32                  `json:"endLedger,omitempty"`
	Filters     []SimpleEventFilter     `json:"filters"`
	Pagination  *SimpleEventsPagination `json:"pagination,omitempty"`
}

type SimpleEventFilter struct {
	Type        string   `json:"type"` // "contract"
	ContractIds []string `json:"contractIds"`
}

type SimpleEventsPagination struct {
	Limit  uint32 `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

type SimpleEventsResponse struct {
	Events       []SimpleContractEvent `json:"events"`
	LatestLedger uint32                `json:"latestLedger"`
	Cursor       string                `json:"cursor"` // necesario para paginar
}

type SimpleContractEvent struct {
	Type                     string   `json:"type"`
	Ledger                   uint32   `json:"ledger"`
	LedgerClosedAt           string   `json:"ledgerClosedAt"`
	ContractId               string   `json:"contractId"`
	ID                       string   `json:"id"`
	OperationIndex           int      `json:"operationIndex"`
	TransactionIndex         int      `json:"transactionIndex"`
	Topic                    []string `json:"topic"`
	Value                    string   `json:"value"` // XDR base64 (string)
	InSuccessfulContractCall bool     `json:"inSuccessfulContractCall"`
	TxHash                   string   `json:"txHash"`
}

// ----------- API pública -----------

// FetchDeposits: trae eventos y mapea solo depósitos válidos (sin placeholders)
func (c *simpleStellarRPCClient) FetchDeposits(ctx context.Context, contractID string) ([]DepositEvent, error) {
	latestLedger, err := c.getLatestLedger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest ledger: %w", err)
	}

	// Ventana configurable (por ahora fija a 10k); evita colapsar a un solo ledger
	const defaultLookback = uint32(10000)
	var startLedger uint32
	if latestLedger > defaultLookback {
		startLedger = latestLedger - defaultLookback
	} else {
		startLedger = 1
	}

	events, err := c.getContractEvents(ctx, contractID, startLedger, latestLedger)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract events: %w", err)
	}

	out := make([]DepositEvent, 0, len(events))
	for _, ev := range events {
		if dep := c.parseDepositEventSimple(ev); dep != nil {
			out = append(out, *dep)
		}
	}
	return out, nil
}

// ----------- Helpers internos -----------

func (c *simpleStellarRPCClient) getLatestLedger(ctx context.Context) (uint32, error) {
	req := SimpleRPCRequest{JSONRPC: "2.0", Method: "getLatestLedger", ID: 1}

	resp, err := c.makeSimpleRPCCall(ctx, req)
	if err != nil {
		return 0, err
	}
	if resp.Error != nil {
		return 0, fmt.Errorf("RPC error: %s", resp.Error.Message)
	}

	var result struct {
		Sequence uint32 `json:"sequence"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return 0, fmt.Errorf("failed to parse latest ledger response: %w", err)
	}
	return result.Sequence, nil
}

func (c *simpleStellarRPCClient) getContractEvents(ctx context.Context, contractID string, startLedger, endLedger uint32) ([]SimpleContractEvent, error) {
	out := make([]SimpleContractEvent, 0, 256)

	cursor := ""
	const perPage uint32 = 100
	const hardCap = 5000

	for {
		params := map[string]any{
			"filters": []map[string]any{
				{"type": "contract", "contractIds": []string{contractID}},
			},
			"pagination": map[string]any{
				"limit": perPage,
			},
		}
		// Solo en la primera página mandamos rango de ledgers.
		if cursor == "" {
			params["startLedger"] = startLedger
			if endLedger > 0 {
				params["endLedger"] = endLedger
			}
		} else {
			// En páginas siguientes, solo cursor (sin start/end).
			params["pagination"].(map[string]any)["cursor"] = cursor
		}

		req := SimpleRPCRequest{
			JSONRPC: "2.0",
			Method:  "getEvents",
			Params:  params,
			ID:      2,
		}

		resp, err := c.makeSimpleRPCCall(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", resp.Error.Message)
		}

		var evResp SimpleEventsResponse
		if err := json.Unmarshal(resp.Result, &evResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events response: %w", err)
		}

		out = append(out, evResp.Events...)
		if evResp.Cursor == "" || len(out) >= hardCap {
			break
		}
		cursor = evResp.Cursor
	}

	if len(out) > hardCap {
		out = out[:hardCap]
	}
	return out, nil
}

func (c *simpleStellarRPCClient) parseDepositEventSimple(event SimpleContractEvent) *DepositEvent {
	// Solo eventos de invocación exitosa
	if !event.InSuccessfulContractCall {
		return nil
	}

	// Depositor: estrictamente cuentas G... (56 chars). Sin fallback inventado.
	depositor := c.extractDepositorSimple(event)
	if !c.looksLikeStellarAccount(depositor) {
		return nil
	}

	// Amount: desactivado hasta tener parser real (XDR/Topic). Sin placeholders.
	amount := c.extractAmountSimple(event)
	if amount == "" || amount == "0" {
		return nil
	}

	return &DepositEvent{
		ContractID:     event.ContractId,
		Depositor:      depositor,
		AmountRaw:      amount,
		OccurredAtUnix: c.parseTimestampSimple(event.LedgerClosedAt),
		ExternalID:     fmt.Sprintf("%s#%d", event.TxHash, event.OperationIndex), // txHash#opIndex
		TxHash:         event.TxHash,
		LedgerSequence: int64(event.Ledger),
		OpIndex:        int32(event.OperationIndex),
		Metadata: map[string]any{
			"source":            "stellar_rpc_simple",
			"event_id":          event.ID,
			"operation_index":   event.OperationIndex,
			"transaction_index": event.TransactionIndex,
			"topics":            event.Topic,
			"xdr_value":         event.Value,
			"event_type":        event.Type,
		},
	}
}

func (c *simpleStellarRPCClient) looksLikeStellarAccount(addr string) bool {
	return len(addr) == 56 && strings.HasPrefix(addr, "G")
}

func (c *simpleStellarRPCClient) extractDepositorSimple(event SimpleContractEvent) string {
	// 1) probar en los topics
	for _, topic := range event.Topic {
		if c.looksLikeStellarAccount(topic) {
			return topic
		}
	}
	// 2) en el futuro, decodificar XDR (event.Value) y validar G...
	if depositor := c.extractAddressFromXDRSimple(event.Value); c.looksLikeStellarAccount(depositor) {
		return depositor
	}
	// 3) sin depositor real, no insertamos
	return ""
}

// Por ahora, no devolvemos montos inventados.
// Implementar parser real cuando definas el esquema (XDR/Topics) del contrato.
func (c *simpleStellarRPCClient) extractAmountSimple(event SimpleContractEvent) string {
	return ""
}

func (c *simpleStellarRPCClient) extractAddressFromXDRSimple(_ string) string {
	// Placeholder: sin implementación (requiere decodificar XDR)
	return ""
}

func (c *simpleStellarRPCClient) parseTimestampSimple(ts string) int64 {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t.Unix()
	}
	return time.Now().Unix()
}

func (c *simpleStellarRPCClient) makeSimpleRPCCall(ctx context.Context, req SimpleRPCRequest) (*SimpleRPCResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(body))
	}

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var resp SimpleRPCResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}
