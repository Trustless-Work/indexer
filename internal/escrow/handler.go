package escrow

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/single", h.postSingle)
	r.Post("/multi", h.postMulti)
	r.Get("/{contractId}", h.getByContract)
	r.Put("/{contractId}", h.putByContract) // reusa upsert
	r.Delete("/{contractId}", h.deleteByContract)
	return r
}

func (h *Handler) postSingle(w http.ResponseWriter, r *http.Request) {
	var in SingleReleaseJSON
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := h.svc.UpsertSingle(r.Context(), in); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "contractId": in.ContractID})
}
func (h *Handler) postMulti(w http.ResponseWriter, r *http.Request) {
	var in MultiReleaseJSON
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := h.svc.UpsertMulti(r.Context(), in); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "contractId": in.ContractID})
}
func (h *Handler) getByContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "contractId")
	out, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	_ = json.NewEncoder(w).Encode(out)
}
func (h *Handler) putByContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "contractId")
	// Detectamos tipo por presencia de "amount" en el JSON (simple para hoy).
	var probe map[string]any
	if err := json.NewDecoder(r.Body).Decode(&probe); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	b, _ := json.Marshal(probe)
	if _, ok := probe["amount"]; ok {
		var in SingleReleaseJSON
		_ = json.Unmarshal(b, &in)
		in.ContractID = id
		if err := h.svc.UpsertSingle(r.Context(), in); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "contractId": id})
		return
	}
	var in MultiReleaseJSON
	_ = json.Unmarshal(b, &in)
	in.ContractID = id
	if err := h.svc.UpsertMulti(r.Context(), in); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "contractId": id})
}
func (h *Handler) deleteByContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "contractId")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "deleted": id})
}
