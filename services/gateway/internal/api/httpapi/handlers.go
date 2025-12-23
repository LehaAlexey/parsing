package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LehaAlexey/Gateway/internal/pb/history"
	"github.com/LehaAlexey/Gateway/internal/pb/users"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	users   UsersClient
	history HistoryClient
}

type UsersClient interface {
	CreateUser(ctx context.Context, email string, name string) (*users.User, error)
	GetUser(ctx context.Context, id string) (*users.User, error)
	AddURL(ctx context.Context, userID string, url string, intervalSeconds int32) (*users.UserURL, error)
	ListURLs(ctx context.Context, userID string, limit int32) ([]*users.UserURL, error)
}

type HistoryClient interface {
	GetHistory(ctx context.Context, productID string, fromUnix int64, toUnix int64, limit int32) ([]*history.HistoryPoint, error)
}

func New(usersClient UsersClient, historyClient HistoryClient) *Handler {
	return &Handler{users: usersClient, history: historyClient}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Post("/users", h.CreateUser)
	r.Get("/users/{id}", h.GetUser)
	r.Post("/users/{id}/urls", h.AddURL)
	r.Get("/users/{id}/urls", h.ListURLs)
	r.Get("/history", h.GetHistory)
	return r
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.users.CreateUser(r.Context(), req.Email, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.users.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) AddURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		URL                   string `json:"url"`
		PollingIntervalSeconds int32  `json:"polling_interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.users.AddURL(r.Context(), id, req.URL, req.PollingIntervalSeconds)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) ListURLs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	res, err := h.users.ListURLs(r.Context(), id, int32(limit))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimSpace(r.URL.Query().Get("product_id"))
	fromUnix := parseInt64Default(r.URL.Query().Get("from"), 0)
	toUnix := parseInt64Default(r.URL.Query().Get("to"), 0)
	limit := parseIntDefault(r.URL.Query().Get("limit"), 1000)
	if productID == "" || fromUnix == 0 || toUnix == 0 {
		writeError(w, http.StatusBadRequest, "product_id, from, to are required")
		return
	}
	res, err := h.history.GetHistory(r.Context(), productID, fromUnix, toUnix, int32(limit))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func parseIntDefault(raw string, def int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func parseInt64Default(raw string, def int64) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return def
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	if msg == "" {
		msg = "error"
	}
	writeJSON(w, status, map[string]string{"error": msg})
}
