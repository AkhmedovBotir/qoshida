package customer

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
)

const maxAvatarBytes = 10 * 1024 * 1024

func saveAvatar(id uuid.UUID, raw, previous string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return previous, nil
	}
	if strings.HasPrefix(raw, "/uploads/customers/") {
		return raw, nil
	}
	path, err := writeAvatar(id, raw)
	if err != nil {
		return "", err
	}
	if previous != "" && previous != path {
		removeAvatarFile(previous)
	}
	return path, nil
}

func writeAvatar(id uuid.UUID, raw string) (string, error) {
	if !strings.Contains(raw, ";base64,") {
		return "", apperror.BadRequest("Rasm formati noto‘g‘ri")
	}
	mime, payload, ok := strings.Cut(raw, ";base64,")
	if !ok {
		return "", apperror.BadRequest("Rasm formati noto‘g‘ri")
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
		return "", apperror.BadRequest("Rasm faqat PNG, JPG yoki WEBP bo‘lishi mumkin")
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", apperror.BadRequest("Rasmni o‘qib bo‘lmadi")
	}
	if len(data) > maxAvatarBytes {
		return "", apperror.BadRequest("Rasm 10 MB dan oshmasin")
	}
	dir := uploadsDir()
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", apperror.Internal()
	}
	name := id.String() + "_" + uuid.NewString()[:8] + ext
	if err = os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", apperror.Internal()
	}
	return "/uploads/customers/" + name, nil
}

func removeAvatarFile(path string) {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/uploads/customers/") {
		return
	}
	_ = os.Remove(filepath.Join(uploadsDir(), filepath.Base(path)))
}

func uploadsDir() string {
	candidates := []string{
		filepath.Join("uploads", "customers"),
		filepath.Join("1. Backend", "uploads", "customers"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append([]string{filepath.Join(wd, "uploads", "customers")}, candidates...)
	}
	for _, dir := range candidates {
		parent := filepath.Dir(dir)
		if _, err := os.Stat(parent); err == nil {
			return dir
		}
	}
	return filepath.Join("uploads", "customers")
}
