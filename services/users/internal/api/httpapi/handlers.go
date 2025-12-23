package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LehaAlexey/Users/internal/models"
	"github.com/LehaAlexey/Users/internal/services/userservice"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	CreateUser(ctx context.Context, req userservice.CreateUserRequest) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	AddURL(ctx context.Context, userID string, rawURL string, intervalSeconds int) (*models.UserURL, error)
	ListUserURLs(ctx context.Context, userID string, limit int) ([]models.UserURL, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Post("/users", h.CreateUser)
	r.Get("/users/{id}", h.GetUser)
	r.Post("/users/{id}/urls", h.AddURL)
	r.Get("/users/{id}/urls", h.ListUserURLs)
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
	res, err := h.service.CreateUser(r.Context(), userservice.CreateUserRequest{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.service.GetUser(r.Context(), id)
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
		PollingIntervalSeconds int    `json:"polling_interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.service.AddURL(r.Context(), id, req.URL, req.PollingIntervalSeconds)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *Handler) ListUserURLs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	if limit <= 0 {
		limit = 100
	}
	res, err := h.service.ListUserURLs(r.Context(), id, limit)
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
