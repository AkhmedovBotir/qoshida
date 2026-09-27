package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"qoshida/backend/internal/modules/admin"
	"qoshida/backend/internal/shared/audit"
	"qoshida/backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	user, access, refresh, err := h.service.Login(r.Context(), req)
	if err != nil {
		response.Error(w, err)
		return
	}
	h.service.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.service.Logout(r.Context(), RefreshFromRequest(r))
	h.service.ClearCookies(w)
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	user, access, refresh, err := h.service.Refresh(r.Context(), RefreshFromRequest(r))
	if err != nil {
		response.Error(w, err)
		return
	}
	h.service.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	id, access, refresh, err := h.service.Restore(r.Context(), AccessFromRequest(r), RefreshFromRequest(r))
	if err != nil {
		response.Error(w, err)
		return
	}
	if access != "" {
		h.service.SetCookies(w, access, refresh)
	}
	user, err := h.service.Me(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, access, refresh, err := h.service.Restore(r.Context(), AccessFromRequest(r), RefreshFromRequest(r))
		if err != nil {
			response.Error(w, err)
			return
		}
		if access != "" {
			h.service.SetCookies(w, access, refresh)
		}
		user, err := h.service.Me(r.Context(), id)
		if err != nil {
			response.Error(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), admin.ActorIDKey, id)
		ctx = audit.With(ctx, audit.Actor{
			Type: audit.TypeAdmin,
			ID:   id,
			Name: strings.TrimSpace(user.FirstName + " " + user.LastName),
			Role: audit.RoleAdmin(),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ActorID(r *http.Request) uuid.UUID {
	id, _ := r.Context().Value(admin.ActorIDKey).(uuid.UUID)
	return id
}

func (h *Handler) Routes(authRPM int) chi.Router {
	r := chi.NewRouter()
	r.Group(func(gr chi.Router) {
		gr.Use(httprate.LimitByIP(authRPM, time.Minute))
		gr.Post("/login", h.Login)
	})
	r.Post("/logout", h.Logout)
	r.Post("/refresh", h.Refresh)
	r.Get("/me", h.Me)
	return r
}
