package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Trustless-Work/indexer/internal/deposits"
	"github.com/Trustless-Work/indexer/internal/escrow"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseLimit(q string, def int) int {
	if q == "" {
		return def
	}
	if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 500 {
		return n
	}
	return def
}

func New(addr string, escrowSvc *escrow.Service, depSvc *deposits.Service) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.Timeout(30*time.Second))

	// Health JSON
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	// Escrows CRUD (nota: /escrows puede no tener GET "/" si no lo definiste en el handler)
	eh := escrow.NewHandler(escrowSvc)
	r.Mount("/escrows", eh.Routes())

	// ---- Depósitos ----

	// Indexación manual (POST) — devuelve JSON siempre (ok/err)
	r.Post("/index/funder-deposits/{contractId}", func(w http.ResponseWriter, req *http.Request) {
		contractID := chi.URLParam(req, "contractId")
		if contractID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing contractId"})
			return
		}
		out, err := depSvc.IndexContractDeposits(req.Context(), contractID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": len(out), "deposits": out})
	})

	// Exploración rápida: lista depósitos por contrato
	r.Get("/deposits/{contractId}", func(w http.ResponseWriter, req *http.Request) {
		contractID := chi.URLParam(req, "contractId")
		if contractID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing contractId"})
			return
		}
		limit := parseLimit(req.URL.Query().Get("limit"), 50)
		list, err := depSvc.Repository().ListByContract(req.Context(), contractID, limit)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, list)
	})

	return &http.Server{Addr: addr, Handler: r}
}
