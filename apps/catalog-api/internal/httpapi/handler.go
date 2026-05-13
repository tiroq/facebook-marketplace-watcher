package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tiroq/fb-market-watcher/catalog-api/internal/repository"
	"github.com/tiroq/fb-market-watcher/catalog-api/internal/service"
)

// Handler holds HTTP handler dependencies.
type Handler struct {
	svc  *service.Service
	repo *repository.Repository
}

// New creates a new Handler.
func New(svc *service.Service, repo *repository.Repository) *Handler {
	return &Handler{svc: svc, repo: repo}
}

// Routes returns an http.Handler with all routes registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("GET /readyz", h.handleReadyz)
	mux.HandleFunc("GET /v1/listings", h.handleListListings)
	mux.HandleFunc("GET /v1/listings/{id}", h.handleGetListing)
	mux.HandleFunc("POST /v1/listings/observations", h.handleCreateObservation)
	mux.HandleFunc("GET /v1/search-queries", h.handleListSearchQueries)
	mux.HandleFunc("POST /v1/search-queries", h.handleCreateSearchQuery)
	return mux
}

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func (h *Handler) handleListListings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	listings, err := h.repo.ListListings(r.Context(), repository.ListListingsFilter{
		Limit:  limit,
		Offset: offset,
		Source: q.Get("source"),
		Status: q.Get("status"),
		Query:  q.Get("q"),
	})
	if err != nil {
		slog.Error("list listings error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if listings == nil {
		listings = []repository.Listing{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"listings": listings})
}

func (h *Handler) handleGetListing(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	listing, err := h.repo.GetListing(r.Context(), id)
	if err != nil {
		slog.Error("get listing error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if listing == nil {
		writeError(w, http.StatusNotFound, "listing not found")
		return
	}
	writeJSON(w, http.StatusOK, listing)
}

func (h *Handler) handleCreateObservation(w http.ResponseWriter, r *http.Request) {
	var req service.ObservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.HandleObservation(r.Context(), req)
	if err != nil {
		slog.Error("handle observation error", "error", err)
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *Handler) handleListSearchQueries(w http.ResponseWriter, r *http.Request) {
	queries, err := h.repo.ListSearchQueries(r.Context())
	if err != nil {
		slog.Error("list search queries error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if queries == nil {
		queries = []repository.SearchQuery{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"search_queries": queries})
}

func (h *Handler) handleCreateSearchQuery(w http.ResponseWriter, r *http.Request) {
	var input repository.CreateSearchQueryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name == "" || input.Query == "" {
		writeError(w, http.StatusBadRequest, "name and query are required")
		return
	}

	q, err := h.repo.CreateSearchQuery(r.Context(), input)
	if err != nil {
		slog.Error("create search query error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, q)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
