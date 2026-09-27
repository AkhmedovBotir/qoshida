package product

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

func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.adminList)
	r.Post("/", h.adminCreate)
	r.Post("/{id}/approve", h.adminApprove)
	r.Post("/{id}/reject", h.adminReject)
	r.Get("/{id}", h.adminGet)
	r.Put("/{id}", h.adminUpdate)
	r.Delete("/{id}", h.adminDelete)
	return r
}

func (h *Handler) adminList(w http.ResponseWriter, r *http.Request)    { h.respondList(w, r, nil) }
func (h *Handler) adminGet(w http.ResponseWriter, r *http.Request)     { h.respondGet(w, r, nil) }
func (h *Handler) adminCreate(w http.ResponseWriter, r *http.Request)  { h.respondCreate(w, r, nil) }
func (h *Handler) adminUpdate(w http.ResponseWriter, r *http.Request)  { h.respondUpdate(w, r, nil) }
func (h *Handler) adminDelete(w http.ResponseWriter, r *http.Request)  { h.respondDelete(w, r, nil) }
func (h *Handler) adminApprove(w http.ResponseWriter, r *http.Request) { h.respondApprove(w, r, nil) }
func (h *Handler) adminReject(w http.ResponseWriter, r *http.Request)  { h.respondReject(w, r, nil) }

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
func (h *Handler) ScopedApprove(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondApprove(w, r, scope)
}
func (h *Handler) ScopedReject(w http.ResponseWriter, r *http.Request, scope *Scope) {
	h.respondReject(w, r, scope)
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
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
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
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
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
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	if err = h.service.Delete(r.Context(), id, scope); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) respondApprove(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Approve(r.Context(), id, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func (h *Handler) respondReject(w http.ResponseWriter, r *http.Request, scope *Scope) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	var req RejectRequest
	if err = response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	item, err := h.service.Reject(r.Context(), id, req.Note, scope)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, item)
}

func listQueryFrom(r *http.Request) ListQuery {
	q := ListQuery{
		Query:          strings.TrimSpace(r.URL.Query().Get("q")),
		ApprovalStatus: strings.TrimSpace(r.URL.Query().Get("approval_status")),
		Page:           atoi(r.URL.Query().Get("page"), 1),
		Limit:          atoi(r.URL.Query().Get("limit"), 20),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("kontragent_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.KontragentID = &id
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("category_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.CategoryID = &id
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
