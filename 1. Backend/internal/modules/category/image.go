package category

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func saveDataImage(sourceID, raw string, maxBytes int) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	mime, payload, ok := strings.Cut(raw, ";base64,")
	if !ok {
		return "", fmt.Errorf("rasm formati noto'g'ri")
	}

	ext := ".png"
	switch {
	case strings.Contains(mime, "image/jpeg"), strings.Contains(mime, "image/jpg"):
		ext = ".jpg"
	case strings.Contains(mime, "image/webp"):
		ext = ".webp"
	case strings.Contains(mime, "image/gif"):
		ext = ".gif"
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("rasm decode: %w", err)
	}
	if maxBytes > 0 && len(data) > maxBytes {
		if maxBytes >= 1024*1024 {
			return "", fmt.Errorf("Rasm %d MB dan oshmasin", maxBytes/(1024*1024))
		}
		return "", fmt.Errorf("Rasm %d KB dan oshmasin", maxBytes/1024)
	}

	dir := uploadsDir()
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	name := sanitizeFileName(sourceID) + ext
	if err = os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", err
	}
	return "/uploads/categories/" + name, nil
}

func uploadsDir() string {
	candidates := []string{
		filepath.Join("uploads", "categories"),
		filepath.Join("1. Backend", "uploads", "categories"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append([]string{filepath.Join(wd, "uploads", "categories")}, candidates...)
	}
	for _, dir := range candidates {
		parent := filepath.Dir(dir)
		if _, err := os.Stat(parent); err == nil {
			return dir
		}
	}
	return filepath.Join("uploads", "categories")
}

func sanitizeFileName(id string) string {
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" {
		return "category"
	}
	return s
}
