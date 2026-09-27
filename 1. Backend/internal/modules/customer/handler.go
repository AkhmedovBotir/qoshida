package customer

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"qoshida/backend/internal/shared/audit"
	"qoshida/backend/internal/shared/response"
)

type contextKey string

const ActorKey contextKey = "customer"

type Handler struct {
	service *Service
	auth    *AuthService
}

func NewHandler(service *Service, auth *AuthService) *Handler {
	return &Handler{service: service, auth: auth}
}

func (h *Handler) AuthRoutes(authRPM int) chi.Router {
	r := chi.NewRouter()
	r.Group(func(gr chi.Router) {
		gr.Use(httprate.LimitByIP(authRPM, time.Minute))
		gr.Post("/check-phone", h.checkPhone)
		gr.Post("/send-code", h.sendCode)
		gr.Post("/verify-code", h.verifyCode)
		gr.Post("/resend-code", h.sendCode)
		gr.Post("/register", h.register)
		gr.Post("/set-password", h.setPassword)
		gr.Post("/login", h.login)
	})
	r.Post("/logout", h.logout)
	r.Post("/refresh", h.refresh)
	r.Get("/me", h.me)
	return r
}

func (h *Handler) AppRoutes() chi.Router {
	r := chi.NewRouter()
	h.RegisterApp(r)
	return r
}

func (h *Handler) RegisterApp(r chi.Router) {
	r.Get("/profile", h.profile)
	r.Put("/profile", h.updateProfile)
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, access, refresh, err := h.auth.Restore(r.Context(), AccessFromRequest(r), RefreshFromRequest(r))
		if err != nil {
			response.Error(w, err)
			return
		}
		if access != "" {
			h.auth.SetCookies(w, access, refresh)
		}
		item, err := h.auth.Actor(r.Context(), id)
		if err != nil {
			response.Error(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), ActorKey, item)
		ctx = audit.With(ctx, audit.Actor{
			Type: "customer",
			ID:   item.ID,
			Name: item.Name,
			Role: "Mijoz",
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func actorFrom(r *http.Request) Customer {
	item, _ := r.Context().Value(ActorKey).(Customer)
	return item
}

func (h *Handler) checkPhone(w http.ResponseWriter, r *http.Request) {
	var req PhoneRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.auth.CheckPhone(r.Context(), req.Phone)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) sendCode(w http.ResponseWriter, r *http.Request) {
	var req PhoneRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.auth.SendCode(r.Context(), req.Phone, req.Purpose)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) verifyCode(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	result, access, refresh, err := h.auth.VerifyCode(r.Context(), req.Phone, req.Code, req.Purpose)
	if err != nil {
		response.Error(w, err)
		return
	}
	if access != "" {
		h.auth.SetCookies(w, access, refresh)
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	fields, err := h.service.ParseProfile(r.Context(), ProfileRequest{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		BirthDate:  req.BirthDate,
		Name:       req.Name,
		RegionID:   req.RegionID,
		DistrictID: req.DistrictID,
		MFYID:      req.MFYID,
	}, true)
	if err != nil {
		response.Error(w, err)
		return
	}
	item, access, refresh, err := h.auth.Register(r.Context(), req.Phone, fields)
	if err != nil {
		response.Error(w, err)
		return
	}
	h.auth.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusCreated, AuthResult{Customer: item.Public()})
}

func (h *Handler) setPassword(w http.ResponseWriter, r *http.Request) {
	var req SetPasswordRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	purpose := strings.TrimSpace(r.URL.Query().Get("purpose"))
	if purpose == "" {
		purpose = PurposeSetup
	}
	item, access, refresh, err := h.auth.SetPassword(r.Context(), req.Phone, req.Password, purpose)
	if err != nil {
		response.Error(w, err)
		return
	}
	h.auth.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusOK, AuthResult{Customer: item.Public()})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, access, refresh, err := h.auth.Login(r.Context(), req.Phone, req.Password)
	if err != nil {
		response.Error(w, err)
		return
	}
	h.auth.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusOK, AuthResult{Customer: item.Public()})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	_ = h.auth.Logout(r.Context(), RefreshFromRequest(r))
	h.auth.ClearCookies(w)
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	item, access, refresh, err := h.auth.Refresh(r.Context(), RefreshFromRequest(r))
	if err != nil {
		response.Error(w, err)
		return
	}
	h.auth.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id, access, refresh, err := h.auth.Restore(r.Context(), AccessFromRequest(r), RefreshFromRequest(r))
	if err != nil {
		response.Error(w, err)
		return
	}
	if access != "" {
		h.auth.SetCookies(w, access, refresh)
	}
	item, err := h.auth.Me(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, actorFrom(r).Public())
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var req ProfileRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.UpdateProfile(r.Context(), actorFrom(r).ID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}
