package shopstock

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
	"qoshida/backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) AdminProductRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.adminList)
	r.Post("/", h.adminCreate)
	r.Get("/{id}", h.adminGet)
	r.Put("/{id}", h.adminUpdate)
	r.Delete("/{id}", h.adminDelete)
	return r
}

func (h *Handler) AdminIncomingRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.adminIncomingList)
	r.Post("/", h.adminIncomingCreate)
	return r
}

func (h *Handler) adminList(w http.ResponseWriter, r *http.Request)     { h.respondList(w, r, nil) }
func (h *Handler) adminGet(w http.ResponseWriter, r *http.Request)      { h.respondGet(w, r, nil) }
func (h *Handler) adminCreate(w http.ResponseWriter, r *http.Request)   { h.respondCreate(w, r, nil) }
func (h *Handler) adminUpdate(w http.ResponseWriter, r *http.Request)   { h.respondUpdate(w, r, nil) }
func (h *Handler) adminDelete(w http.ResponseWriter, r *http.Request)   { h.respondDelete(w, r, nil) }
func (h *Handler) adminIncomingList(w http.ResponseWriter, r *http.Request) {
	h.respondIncomingList(w, r, nil)
}
func (h *Handler) adminIncomingCreate(w http.ResponseWriter, r *http.Request) {
	h.respondIncomingCreate(w, r, nil)
}

func (h *Handler) ScopedList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondList(w, r, scope)
}
func (h *Handler) ScopedGet(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondGet(w, r, scope)
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
func (h *Handler) ScopedIncomingList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondIncomingList(w, r, scope)
}
func (h *Handler) ScopedIncomingCreate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondIncomingCreate(w, r, scope)
}

func (h *Handler) respondList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	result, err := h.service.ListProducts(r.Context(), productQueryFrom(r), scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) respondGet(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.GetProduct(r.Context(), id, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondCreate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	var req AttachRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Attach(r.Context(), req, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func (h *Handler) respondUpdate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	var req UpdateRequest
	if err = response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.UpdateProduct(r.Context(), id, req, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondDelete(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	if err = h.service.DeleteProduct(r.Context(), id, scope); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) respondIncomingList(w http.ResponseWriter, r *http.Request, scope *Scope) {
	result, err := h.service.ListIncomings(r.Context(), incomingQueryFrom(r), scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) respondIncomingCreate(w http.ResponseWriter, r *http.Request, scope *Scope) {
	var req IncomingRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.CreateIncoming(r.Context(), req, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, item)
}

func productQueryFrom(r *http.Request) ProductListQuery {
	q := ProductListQuery{
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Page:  atoi(r.URL.Query().Get("page"), 1),
		Limit: atoi(r.URL.Query().Get("limit"), 20),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("shop_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ShopID = &id
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("template_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.TemplateID = &id
		}
	}
	return q
}

func incomingQueryFrom(r *http.Request) IncomingListQuery {
	q := IncomingListQuery{
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
		Page:  atoi(r.URL.Query().Get("page"), 1),
		Limit: atoi(r.URL.Query().Get("limit"), 20),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("shop_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ShopID = &id
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("shop_product_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ShopProductID = &id
		}
	}
	return q
}

func parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperror.BadRequest("ID noto'g'ri")
	}
	return id, nil
}

func atoi(raw string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return n
}
