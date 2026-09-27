package market

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"qoshida/backend/internal/modules/customer"
	"qoshida/backend/internal/modules/region"
	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/response"
)

type Handler struct {
	service *Service
	regions *region.Service
}

func NewHandler(service *Service, regions *region.Service) *Handler {
	return &Handler{service: service, regions: regions}
}

func (h *Handler) PublicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/regions", h.regionsList)
	r.Get("/categories", h.categories)
	r.Get("/catalog", h.catalogList)
	r.Get("/catalog/{kind}/{id}", h.catalogGet)
	return r
}

func (h *Handler) AppRoutes() chi.Router {
	r := chi.NewRouter()
	h.RegisterApp(r)
	return r
}

func (h *Handler) RegisterApp(r chi.Router) {
	r.Get("/cart", h.cart)
	r.Put("/cart", h.cartUpsert)
	r.Post("/checkout", h.checkout)
	r.Get("/orders", h.orders)
	r.Get("/orders/{id}", h.orderGet)
}

func (h *Handler) regionsList(w http.ResponseWriter, r *http.Request) {
	q := region.ListQuery{
		Type:  strings.TrimSpace(r.URL.Query().Get("type")),
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Page:  atoi(r.URL.Query().Get("page"), 1),
		Limit: atoi(r.URL.Query().Get("limit"), 100),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("parent_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, apperror.BadRequest("parent_id noto‘g‘ri"))
			return
		}
		q.ParentID = &id
	}
	result, err := h.regions.List(r.Context(), q)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) categories(w http.ResponseWriter, r *http.Request) {
	var parent *uuid.UUID
	if raw := strings.TrimSpace(r.URL.Query().Get("parent_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, apperror.BadRequest("parent_id noto‘g‘ri"))
			return
		}
		parent = &id
	}
	roots := r.URL.Query().Get("roots") == "1" || parent == nil && r.URL.Query().Get("parent_id") == ""
	if r.URL.Query().Get("roots") == "0" {
		roots = false
	}
	items, err := h.service.Categories(r.Context(), parent, roots)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) catalogList(w http.ResponseWriter, r *http.Request) {
	q := ListQuery{
		Kind:  strings.TrimSpace(r.URL.Query().Get("kind")),
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Page:  atoi(r.URL.Query().Get("page"), 1),
		Limit: atoi(r.URL.Query().Get("limit"), 20),
	}
	q.CategoryID = parseQueryUUID(r.URL.Query().Get("category_id"))
	q.RegionID = parseQueryUUID(r.URL.Query().Get("region_id"))
	q.DistrictID = parseQueryUUID(r.URL.Query().Get("district_id"))
	q.MFYID = parseQueryUUID(r.URL.Query().Get("mfy_id"))
	result, err := h.service.ListCatalog(r.Context(), q)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) catalogGet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto‘g‘ri"))
		return
	}
	item, err := h.service.GetCatalog(r.Context(), chi.URLParam(r, "kind"), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) cart(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Cart(r.Context(), actorID(r))
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) cartUpsert(w http.ResponseWriter, r *http.Request) {
	var req CartUpsertRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.service.UpsertCart(r.Context(), actorID(r), req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Checkout(r.Context(), actorID(r), req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) orders(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Orders(r.Context(), actorID(r), atoi(r.URL.Query().Get("page"), 1), atoi(r.URL.Query().Get("limit"), 20))
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) orderGet(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID noto‘g‘ri"))
		return
	}
	item, err := h.service.Order(r.Context(), actorID(r), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func actorID(r *http.Request) uuid.UUID {
	item, _ := r.Context().Value(customer.ActorKey).(customer.Customer)
	return item.ID
}

func parseQueryUUID(raw string) *uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func atoi(raw string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return n
}
