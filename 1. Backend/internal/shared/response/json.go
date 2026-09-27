package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"qoshida/backend/internal/shared/apperror"
)

type Body struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, Body{Success: status < 400, Data: data})
}

func Error(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		write(w, appErr.Code, Body{Success: false, Error: appErr.Message})
		return
	}

	slog.Error("unhandled error", "err", err)
	write(w, http.StatusInternalServerError, Body{Success: false, Error: "Ichki xatolik yuz berdi"})
}

func write(w http.ResponseWriter, status int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func Decode[T any](r *http.Request, dst *T) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return apperror.BadRequest("JSON formati noto'g'ri")
	}
	return nil
}
