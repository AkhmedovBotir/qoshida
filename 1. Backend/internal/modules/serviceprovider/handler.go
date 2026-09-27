package serviceprovider

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"qoshida/backend/internal/modules/activitytype"
	"qoshida/backend/internal/modules/providerservice"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
	"qoshida/backend/internal/shared/response"
)

type contextKey string

const ActorKey contextKey = "service_provider"

type Handler struct {
	service    *Service
	auth       *AuthService
	activities *activitytype.Service
	offerings  *providerservice.Handler
}

func NewHandler(service *Service, auth *AuthService, activities *activitytype.Service) *Handler {
	return &Handler{service: service, auth: auth, activities: activities}
}

func (h *Handler) SetServices(offerings *providerservice.Handler) {
	h.offerings = offerings
}

func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.adminList)
	r.Post("/", h.adminCreate)
	r.Get("/stats", h.adminStats)
	r.Get("/identifications", h.adminIdentList)
	r.Get("/identifications/{id}", h.adminIdentGet)
	r.Post("/identifications/{id}/approve", h.adminIdentApprove)
	r.Post("/identifications/{id}/reject", h.adminIdentReject)
	r.Get("/{id}", h.adminGet)
	r.Put("/{id}", h.adminUpdate)
	r.Delete("/{id}", h.adminDelete)
	return r
}

func (h *Handler) AuthRoutes(authRPM int) chi.Router {
	r := chi.NewRouter()
	r.Group(func(gr chi.Router) {
		gr.Use(httprate.LimitByIP(authRPM, time.Minute))
		gr.Post("/check-phone", h.checkPhone)
		gr.Post("/send-code", h.sendCode)
		gr.Post("/verify-code", h.verifyCode)
		gr.Post("/resend-code", h.sendCode)
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
	r.Get("/profile", h.profile)
	r.Get("/identification", h.ownIdentGet)
	r.Put("/identification", h.ownIdentSubmit)
	r.Get("/services", h.serviceList)
	r.Post("/services", h.serviceCreate)
	r.Put("/services/{id}", h.serviceUpdate)
	r.Delete("/services/{id}", h.serviceDelete)
	return r
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
			Type: audit.TypeServiceProvider,
			ID:   item.ID,
			Name: item.Name,
			Role: audit.RoleServiceProvider(),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func actorFrom(r *http.Request) Provider {
	item, _ := r.Context().Value(ActorKey).(Provider)
	return item
}

func (h *Handler) adminList(w http.ResponseWriter, r *http.Request)  { h.respondList(w, r, nil) }
func (h *Handler) adminGet(w http.ResponseWriter, r *http.Request)   { h.respondGet(w, r, nil) }
func (h *Handler) adminCreate(w http.ResponseWriter, r *http.Request) { h.respondCreate(w, r, nil) }
func (h *Handler) adminUpdate(w http.ResponseWriter, r *http.Request) { h.respondUpdate(w, r, nil) }
func (h *Handler) adminDelete(w http.ResponseWriter, r *http.Request) { h.respondDelete(w, r, nil) }

func (h *Handler) adminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

func (h *Handler) ScopedList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondList(w, r, scope)
}
func (h *Handler) ScopedCreate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondCreate(w, r, scope)
}
func (h *Handler) ScopedUpdate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondUpdate(w, r, scope)
}
func (h *Handler) ScopedDelete(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondDelete(w, r, scope)
}

func (h *Handler) ActiveTypes(w http.ResponseWriter, r *http.Request) {
	result, err := h.activities.List(r.Context(), activitytype.ListQuery{
		Status: activitytype.StatusActive, Page: 1, Limit: 200,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result.Items)
}

func (h *Handler) respondList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	result, err := h.service.List(r.Context(), listQueryFrom(r), scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) respondGet(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	item, err := h.service.Get(r.Context(), id, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondCreate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	var req CreateRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Create(r.Context(), req, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) respondUpdate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	var req CreateRequest
	if err = response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Update(r.Context(), id, req, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondDelete(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	if err = h.service.Delete(r.Context(), id, scope); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	if err := h.auth.SendCode(r.Context(), req.Phone); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "sent", "message": "SMS kodi yuborildi. Kod 5 daqiqa amal qiladi"})
}

func (h *Handler) verifyCode(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.auth.VerifyCode(r.Context(), req.Phone, req.Code); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func (h *Handler) setPassword(w http.ResponseWriter, r *http.Request) {
	var req SetPasswordRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, access, refresh, err := h.auth.SetPassword(r.Context(), req.Phone, req.Password)
	if err != nil {
		response.Error(w, err)
		return
	}
	h.auth.SetCookies(w, access, refresh)
	response.JSON(w, http.StatusOK, AuthResult{Provider: item})
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
	response.JSON(w, http.StatusOK, AuthResult{Provider: item})
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

func (h *Handler) serviceScope(r *http.Request) *providerservice.Scope {
	return providerservice.ScopeFromProvider(actorFrom(r).ID)
}

func (h *Handler) serviceList(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedList(w, r, h.serviceScope(r))
}

func (h *Handler) serviceCreate(w http.ResponseWriter, r *http.Request) {
	if !h.requireIdentified(w, r) {
		return
	}
	h.offerings.ScopedCreate(w, r, h.serviceScope(r))
}

func (h *Handler) serviceUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.requireIdentified(w, r) {
		return
	}
	h.offerings.ScopedUpdate(w, r, h.serviceScope(r))
}

func (h *Handler) serviceDelete(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedDelete(w, r, h.serviceScope(r))
}

func (h *Handler) requireIdentified(w http.ResponseWriter, r *http.Request) bool {
	if actorFrom(r).IdentificationStatus != IdentApproved {
		response.Error(w, apperror.Forbidden("Identifikatsiyadan o‘tmaguncha xizmat joylab bo‘lmaydi"))
		return false
	}
	return true
}

func (h *Handler) adminIdentList(w http.ResponseWriter, r *http.Request) {
	h.respondIdentList(w, r, nil)
}
func (h *Handler) adminIdentGet(w http.ResponseWriter, r *http.Request) {
	h.respondIdentGet(w, r, nil)
}
func (h *Handler) adminIdentApprove(w http.ResponseWriter, r *http.Request) {
	h.respondIdentApprove(w, r, nil)
}
func (h *Handler) adminIdentReject(w http.ResponseWriter, r *http.Request) {
	h.respondIdentReject(w, r, nil)
}

func (h *Handler) ScopedIdentList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondIdentList(w, r, scope)
}
func (h *Handler) ScopedIdentGet(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondIdentGet(w, r, scope)
}
func (h *Handler) ScopedIdentApprove(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondIdentApprove(w, r, scope)
}
func (h *Handler) ScopedIdentReject(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondIdentReject(w, r, scope)
}

func (h *Handler) ownIdentGet(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetOwnIdentification(r.Context(), actorFrom(r).ID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) ownIdentSubmit(w http.ResponseWriter, r *http.Request) {
	var req IdentSubmitRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.SubmitIdentification(r.Context(), actorFrom(r).ID, req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondIdentList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	result, err := h.service.ListIdentifications(r.Context(), identListQueryFrom(r), scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) respondIdentGet(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	item, err := h.service.GetIdentification(r.Context(), id, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondIdentApprove(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	item, err := h.service.ApproveIdentification(r.Context(), id, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondIdentReject(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	var req IdentRejectRequest
	if err = response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.RejectIdentification(r.Context(), id, req.Note, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func identListQueryFrom(r *http.Request) IdentListQuery {
	return IdentListQuery{
		Query:  strings.TrimSpace(r.URL.Query().Get("q")),
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Page:   atoi(r.URL.Query().Get("page"), 1),
		Limit:  atoi(r.URL.Query().Get("limit"), 20),
	}
}

func listQueryFrom(r *http.Request) ListQuery {
	q := ListQuery{
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Page:  atoi(r.URL.Query().Get("page"), 1),
		Limit: atoi(r.URL.Query().Get("limit"), 20),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("region_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.RegionID = &id
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("district_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.DistrictID = &id
		}
	}
	return q
}

func atoi(raw string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return n
}
