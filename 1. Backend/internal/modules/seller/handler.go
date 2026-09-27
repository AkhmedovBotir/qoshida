package seller

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"qoshida/backend/internal/modules/shopstock"
	"qoshida/backend/internal/modules/shoptemplate"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
	"qoshida/backend/internal/shared/response"
)

type contextKey string

const ActorKey contextKey = "seller"

type Handler struct {
	service   *Service
	auth      *AuthService
	templates *shoptemplate.Handler
	stock     *shopstock.Handler
}

func NewHandler(service *Service, auth *AuthService) *Handler {
	return &Handler{service: service, auth: auth}
}

func (h *Handler) SetCatalog(templates *shoptemplate.Handler, stock *shopstock.Handler) {
	h.templates = templates
	h.stock = stock
}

func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.adminList)
	r.Post("/", h.adminCreate)
	r.Get("/stats", h.adminStats)
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
	r.Mount("/shop-templates", h.templates.ReadRoutes())
	r.Get("/shop-products", h.shopProductList)
	r.Post("/shop-products", h.shopProductCreate)
	r.Get("/shop-products/{id}", h.shopProductGet)
	r.Put("/shop-products/{id}", h.shopProductUpdate)
	r.Delete("/shop-products/{id}", h.shopProductDelete)
	r.Get("/shop-incomings", h.shopIncomingList)
	r.Post("/shop-incomings", h.shopIncomingCreate)
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
			Type: audit.TypeSeller,
			ID:   item.ID,
			Name: strings.TrimSpace(item.FirstName + " " + item.LastName),
			Role: audit.RoleSeller(),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func actorFrom(r *http.Request) Seller {
	item, _ := r.Context().Value(ActorKey).(Seller)
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
	response.JSON(w, http.StatusOK, AuthResult{Seller: item})
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
	response.JSON(w, http.StatusOK, AuthResult{Seller: item})
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

func (h *Handler) stockScope(r *http.Request) *shopstock.Scope {
	return shopstock.ScopeFromShop(actorFrom(r).ShopID)
}

func (h *Handler) shopProductList(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedList(w, r, h.stockScope(r))
}
func (h *Handler) shopProductGet(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedGet(w, r, h.stockScope(r))
}
func (h *Handler) shopProductCreate(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedCreate(w, r, h.stockScope(r))
}
func (h *Handler) shopProductUpdate(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedUpdate(w, r, h.stockScope(r))
}
func (h *Handler) shopProductDelete(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedDelete(w, r, h.stockScope(r))
}
func (h *Handler) shopIncomingList(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedIncomingList(w, r, h.stockScope(r))
}
func (h *Handler) shopIncomingCreate(w http.ResponseWriter, r *http.Request) {
	h.stock.ScopedIncomingCreate(w, r, h.stockScope(r))
}

func listQueryFrom(r *http.Request) ListQuery {
	q := ListQuery{
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Page:  atoi(r.URL.Query().Get("page"), 1),
		Limit: atoi(r.URL.Query().Get("limit"), 20),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("shop_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ShopID = &id
		}
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
	if raw := strings.TrimSpace(r.URL.Query().Get("mfy_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.MFYID = &id
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
