package health

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"qoshida/backend/internal/shared/response"
)

type Handler struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "qoshida-backend",
	})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":  "degraded",
			"service": "qoshida-backend",
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"status":  "ready",
		"service": "qoshida-backend",
	})
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Live)
	r.Get("/ready", h.Ready)
	return r
}
