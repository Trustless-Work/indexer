package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Trustless-Work/indexer/internal/deposits"
	"github.com/Trustless-Work/indexer/internal/escrow"
)

func New(addr string, escrowSvc *escrow.Service, depSvc *deposits.Service) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	// Escrows CRUD
	eh := escrow.NewHandler(escrowSvc)
	r.Mount("/escrows", eh.Routes())

	// Indexación manual (pruebas)
	r.Post("/index/funder-deposits/{contractId}", func(w http.ResponseWriter, req *http.Request) {
		contractID := chi.URLParam(req, "contractId")
		out, err := depSvc.IndexContractDeposits(req.Context(), contractID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "deposits": out})
	})

	return &http.Server{Addr: addr, Handler: r}
}
