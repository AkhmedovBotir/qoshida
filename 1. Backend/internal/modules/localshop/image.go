package localshop

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
)

const maxImageBytes = 10 * 1024 * 1024

func saveImage(id uuid.UUID, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if !strings.Contains(raw, ";base64,") {
		return "", apperror.BadRequest("Rasm formati noto'g'ri")
	}
	mime, payload, ok := strings.Cut(raw, ";base64,")
	if !ok {
		return "", apperror.BadRequest("Rasm formati noto'g'ri")
	}

	ext := ".png"
	switch {
	case strings.Contains(mime, "image/jpeg"), strings.Contains(mime, "image/jpg"):
		ext = ".jpg"
	case strings.Contains(mime, "image/webp"):
		ext = ".webp"
	case strings.Contains(mime, "image/png"):
		ext = ".png"
	default:
		return "", apperror.BadRequest("Rasm faqat PNG, JPG yoki WEBP bo'lishi mumkin")
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", apperror.BadRequest("Rasmni o'qib bo'lmadi")
	}
	if len(data) > maxImageBytes {
		return "", apperror.BadRequest("Rasm 10 MB dan oshmasin")
	}

	dir := uploadsDir()
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", apperror.Internal()
	}
	name := id.String() + ext
	if err = os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", apperror.Internal()
	}
	return "/uploads/local-shops/" + name, nil
}

func removeImageFile(path string) {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/uploads/local-shops/") {
		return
	}
	name := filepath.Base(path)
	_ = os.Remove(filepath.Join(uploadsDir(), name))
}

func uploadsDir() string {
	candidates := []string{
		filepath.Join("uploads", "local-shops"),
		filepath.Join("1. Backend", "uploads", "local-shops"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append([]string{filepath.Join(wd, "uploads", "local-shops")}, candidates...)
	}
	for _, dir := range candidates {
		parent := filepath.Dir(dir)
		if _, err := os.Stat(parent); err == nil {
			return dir
		}
	}
	return filepath.Join("uploads", "local-shops")
}
