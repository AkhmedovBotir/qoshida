package providerservice

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
)

const (
	maxImages     = 5
	maxImageBytes = 10 * 1024 * 1024
)

func saveImages(serviceID uuid.UUID, incoming []string, previous []string) ([]string, error) {
	if incoming == nil {
		incoming = []string{}
	}
	if len(incoming) > maxImages {
		return nil, apperror.BadRequest("Rasmlar soni 5 tadan oshmasin")
	}

	kept := map[string]bool{}
	out := make([]string, 0, len(incoming))
	for _, raw := range incoming {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "/uploads/services/") {
			kept[raw] = true
			out = append(out, raw)
			continue
		}
		path, err := saveImage(serviceID, raw)
		if err != nil {
			return nil, err
		}
		if path != "" {
			out = append(out, path)
		}
	}
	if len(out) > maxImages {
		return nil, apperror.BadRequest("Rasmlar soni 5 tadan oshmasin")
	}
	if len(out) == 0 {
		return nil, apperror.BadRequest("Kamida bitta rasm yuklang")
	}
	for _, old := range previous {
		if !kept[old] {
			removeImageFile(old)
		}
	}
	return out, nil
}

func saveImage(serviceID uuid.UUID, raw string) (string, error) {
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
	if len(data) > maxImageBytes {
		return "", apperror.BadRequest("Rasm 10 MB dan oshmasin")
	}

	dir := uploadsDir()
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", apperror.Internal()
	}
	name := serviceID.String() + "_" + uuid.NewString()[:8] + ext
	if err = os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", apperror.Internal()
	}
	return "/uploads/services/" + name, nil
}

func removeImageFile(path string) {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/uploads/services/") {
		return
	}
	_ = os.Remove(filepath.Join(uploadsDir(), filepath.Base(path)))
}

func removeAllImages(paths []string) {
	for _, path := range paths {
		removeImageFile(path)
	}
}

func uploadsDir() string {
	candidates := []string{
		filepath.Join("uploads", "services"),
		filepath.Join("1. Backend", "uploads", "services"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append([]string{filepath.Join(wd, "uploads", "services")}, candidates...)
	}
	for _, dir := range candidates {
		parent := filepath.Dir(dir)
		if _, err := os.Stat(parent); err == nil {
			return dir
		}
	}
	return filepath.Join("uploads", "services")
}
