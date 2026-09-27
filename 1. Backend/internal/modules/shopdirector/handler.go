package shopdirector

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"qoshida/backend/internal/modules/category"
	"qoshida/backend/internal/modules/delivery"
	"qoshida/backend/internal/modules/kontragent"
	"qoshida/backend/internal/modules/localshop"
	"qoshida/backend/internal/modules/product"
	"qoshida/backend/internal/modules/providerservice"
	"qoshida/backend/internal/modules/seller"
	"qoshida/backend/internal/modules/serviceprovider"
	"qoshida/backend/internal/modules/shopstock"
	"qoshida/backend/internal/modules/shoptemplate"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/audit"
	"qoshida/backend/internal/shared/response"
)

type contextKey string

const ActorKey contextKey = "shop_director"

type Handler struct {
	service     *Service
	auth        *AuthService
	kontragents *kontragent.Handler
	shops       *localshop.Handler
	providers   *serviceprovider.Handler
	sellers     *seller.Handler
	deliveries  *delivery.Handler
	products    *product.Handler
	categories  *category.Handler
	templates   *shoptemplate.Handler
	stock       *shopstock.Handler
	offerings   *providerservice.Handler
}

func NewHandler(service *Service, auth *AuthService, kontragents *kontragent.Handler, shops *localshop.Handler, providers *serviceprovider.Handler, sellers *seller.Handler, deliveries *delivery.Handler, products *product.Handler, categories *category.Handler, templates *shoptemplate.Handler, stock *shopstock.Handler, offerings *providerservice.Handler) *Handler {
	return &Handler{service: service, auth: auth, kontragents: kontragents, shops: shops, providers: providers, sellers: sellers, deliveries: deliveries, products: products, categories: categories, templates: templates, stock: stock, offerings: offerings}
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
	r.Get("/activity-types", h.kontragents.ActiveTypes)
	r.Get("/kontragents", h.kontragentList)
	r.Post("/kontragents", h.kontragentCreate)
	r.Put("/kontragents/{id}", h.kontragentUpdate)
	r.Delete("/kontragents/{id}", h.kontragentDelete)
	r.Get("/local-shops", h.localShopList)
	r.Post("/local-shops", h.localShopCreate)
	r.Put("/local-shops/{id}", h.localShopUpdate)
	r.Delete("/local-shops/{id}", h.localShopDelete)
	r.Get("/service-providers", h.providerList)
	r.Post("/service-providers", h.providerCreate)
	r.Put("/service-providers/{id}", h.providerUpdate)
	r.Delete("/service-providers/{id}", h.providerDelete)
	r.Get("/provider-identifications", h.identList)
	r.Get("/provider-identifications/{id}", h.identGet)
	r.Post("/provider-identifications/{id}/approve", h.identApprove)
	r.Post("/provider-identifications/{id}/reject", h.identReject)
	r.Get("/provider-services", h.offeringList)
	r.Post("/provider-services", h.offeringCreate)
	r.Post("/provider-services/{id}/approve", h.offeringApprove)
	r.Post("/provider-services/{id}/reject", h.offeringReject)
	r.Put("/provider-services/{id}", h.offeringUpdate)
	r.Delete("/provider-services/{id}", h.offeringDelete)
	r.Get("/sellers", h.sellerList)
	r.Post("/sellers", h.sellerCreate)
	r.Put("/sellers/{id}", h.sellerUpdate)
	r.Delete("/sellers/{id}", h.sellerDelete)
	r.Get("/deliveries", h.deliveryList)
	r.Post("/deliveries", h.deliveryCreate)
	r.Put("/deliveries/{id}", h.deliveryUpdate)
	r.Delete("/deliveries/{id}", h.deliveryDelete)
	r.Get("/products", h.productList)
	r.Post("/products", h.productCreate)
	r.Post("/products/{id}/approve", h.productApprove)
	r.Post("/products/{id}/reject", h.productReject)
	r.Put("/products/{id}", h.productUpdate)
	r.Delete("/products/{id}", h.productDelete)
	r.Mount("/categories", h.categories.AppRoutes())
	r.Mount("/shop-templates", h.templates.AdminRoutes())
	r.Get("/shop-products", h.shopProductList)
	r.Post("/shop-products", h.shopProductCreate)
	r.Get("/shop-products/{id}", h.shopProductGet)
	r.Put("/shop-products/{id}", h.shopProductUpdate)
	r.Delete("/shop-products/{id}", h.shopProductDelete)
	r.Get("/shop-incomings", h.shopIncomingList)
	r.Post("/shop-incomings", h.shopIncomingCreate)
	return r
}

func (h *Handler) kontragentScope(r *http.Request) *kontragent.Scope {
	actor := actorFrom(r)
	return kontragent.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) kontragentList(w http.ResponseWriter, r *http.Request) {
	h.kontragents.ScopedList(w, r, h.kontragentScope(r))
}

func (h *Handler) kontragentCreate(w http.ResponseWriter, r *http.Request) {
	h.kontragents.ScopedCreate(w, r, h.kontragentScope(r))
}

func (h *Handler) kontragentUpdate(w http.ResponseWriter, r *http.Request) {
	h.kontragents.ScopedUpdate(w, r, h.kontragentScope(r))
}

func (h *Handler) kontragentDelete(w http.ResponseWriter, r *http.Request) {
	h.kontragents.ScopedDelete(w, r, h.kontragentScope(r))
}

func (h *Handler) localShopScope(r *http.Request) *localshop.Scope {
	actor := actorFrom(r)
	return localshop.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) localShopList(w http.ResponseWriter, r *http.Request) {
	h.shops.ScopedList(w, r, h.localShopScope(r))
}

func (h *Handler) localShopCreate(w http.ResponseWriter, r *http.Request) {
	h.shops.ScopedCreate(w, r, h.localShopScope(r))
}

func (h *Handler) localShopUpdate(w http.ResponseWriter, r *http.Request) {
	h.shops.ScopedUpdate(w, r, h.localShopScope(r))
}

func (h *Handler) localShopDelete(w http.ResponseWriter, r *http.Request) {
	h.shops.ScopedDelete(w, r, h.localShopScope(r))
}

func (h *Handler) providerScope(r *http.Request) *serviceprovider.Scope {
	actor := actorFrom(r)
	return serviceprovider.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) providerList(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedList(w, r, h.providerScope(r))
}

func (h *Handler) providerCreate(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedCreate(w, r, h.providerScope(r))
}

func (h *Handler) providerUpdate(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedUpdate(w, r, h.providerScope(r))
}

func (h *Handler) providerDelete(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedDelete(w, r, h.providerScope(r))
}

func (h *Handler) identList(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedIdentList(w, r, h.providerScope(r))
}

func (h *Handler) identGet(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedIdentGet(w, r, h.providerScope(r))
}

func (h *Handler) identApprove(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedIdentApprove(w, r, h.providerScope(r))
}

func (h *Handler) identReject(w http.ResponseWriter, r *http.Request) {
	h.providers.ScopedIdentReject(w, r, h.providerScope(r))
}

func (h *Handler) sellerScope(r *http.Request) *seller.Scope {
	actor := actorFrom(r)
	return seller.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) sellerList(w http.ResponseWriter, r *http.Request) {
	h.sellers.ScopedList(w, r, h.sellerScope(r))
}

func (h *Handler) sellerCreate(w http.ResponseWriter, r *http.Request) {
	h.sellers.ScopedCreate(w, r, h.sellerScope(r))
}

func (h *Handler) sellerUpdate(w http.ResponseWriter, r *http.Request) {
	h.sellers.ScopedUpdate(w, r, h.sellerScope(r))
}

func (h *Handler) sellerDelete(w http.ResponseWriter, r *http.Request) {
	h.sellers.ScopedDelete(w, r, h.sellerScope(r))
}

func (h *Handler) deliveryScope(r *http.Request) *delivery.Scope {
	actor := actorFrom(r)
	return delivery.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) deliveryList(w http.ResponseWriter, r *http.Request) {
	h.deliveries.ScopedList(w, r, h.deliveryScope(r))
}

func (h *Handler) deliveryCreate(w http.ResponseWriter, r *http.Request) {
	h.deliveries.ScopedCreate(w, r, h.deliveryScope(r))
}

func (h *Handler) deliveryUpdate(w http.ResponseWriter, r *http.Request) {
	h.deliveries.ScopedUpdate(w, r, h.deliveryScope(r))
}

func (h *Handler) deliveryDelete(w http.ResponseWriter, r *http.Request) {
	h.deliveries.ScopedDelete(w, r, h.deliveryScope(r))
}

func (h *Handler) offeringScope(r *http.Request) *providerservice.Scope {
	actor := actorFrom(r)
	return providerservice.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) offeringList(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedList(w, r, h.offeringScope(r))
}

func (h *Handler) offeringCreate(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedCreate(w, r, h.offeringScope(r))
}

func (h *Handler) offeringUpdate(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedUpdate(w, r, h.offeringScope(r))
}

func (h *Handler) offeringDelete(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedDelete(w, r, h.offeringScope(r))
}

func (h *Handler) offeringApprove(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedApprove(w, r, h.offeringScope(r))
}

func (h *Handler) offeringReject(w http.ResponseWriter, r *http.Request) {
	h.offerings.ScopedReject(w, r, h.offeringScope(r))
}

func (h *Handler) productScope(r *http.Request) *product.Scope {
	actor := actorFrom(r)
	return product.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
}

func (h *Handler) productList(w http.ResponseWriter, r *http.Request) {
	h.products.ScopedList(w, r, h.productScope(r))
}

func (h *Handler) productCreate(w http.ResponseWriter, r *http.Request) {
	h.products.ScopedCreate(w, r, h.productScope(r))
}

func (h *Handler) productUpdate(w http.ResponseWriter, r *http.Request) {
	h.products.ScopedUpdate(w, r, h.productScope(r))
}

func (h *Handler) productDelete(w http.ResponseWriter, r *http.Request) {
	h.products.ScopedDelete(w, r, h.productScope(r))
}

func (h *Handler) productApprove(w http.ResponseWriter, r *http.Request) {
	h.products.ScopedApprove(w, r, h.productScope(r))
}

func (h *Handler) productReject(w http.ResponseWriter, r *http.Request) {
	h.products.ScopedReject(w, r, h.productScope(r))
}

func (h *Handler) stockScope(r *http.Request) *shopstock.Scope {
	actor := actorFrom(r)
	return shopstock.ScopeFromDirector(actor.RegionID, actor.DistrictID, actor.MFYID)
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
			Type: audit.TypeShopDirector,
			ID:   item.ID,
			Name: strings.TrimSpace(item.FirstName + " " + item.LastName),
			Role: audit.RoleDirector(),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func actorFrom(r *http.Request) Director {
	item, _ := r.Context().Value(ActorKey).(Director)
	return item
}

func (h *Handler) adminList(w http.ResponseWriter, r *http.Request) {
	q := listQueryFrom(r)
	result, err := h.service.List(r.Context(), q, nil)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) adminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

func (h *Handler) adminGet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	item, err := h.service.Get(r.Context(), id, nil)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) adminCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Create(r.Context(), req, nil)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) adminUpdate(w http.ResponseWriter, r *http.Request) {
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
	item, err := h.service.Update(r.Context(), id, req, nil)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) adminDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto'g'ri"))
		return
	}
	if err = h.service.Delete(r.Context(), id, nil); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ManagerList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	result, err := h.service.List(r.Context(), listQueryFrom(r), scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) ManagerCreate(w http.ResponseWriter, r *http.Request, scope *Scope) {
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

func (h *Handler) ManagerUpdate(w http.ResponseWriter, r *http.Request, scope *Scope) {
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

func (h *Handler) ManagerDelete(w http.ResponseWriter, r *http.Request, scope *Scope) {
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
	response.JSON(w, http.StatusOK, AuthResult{Director: item})
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
	response.JSON(w, http.StatusOK, AuthResult{Director: item})
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
