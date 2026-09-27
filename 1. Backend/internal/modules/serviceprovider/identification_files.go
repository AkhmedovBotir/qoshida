package serviceprovider

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"qoshida/backend/internal/shared/apperror"
)

const (
	maxIdentImageBytes = 10 * 1024 * 1024
	maxIdentPDFBytes   = 5 * 1024 * 1024
)

func saveIdentImage(providerID uuid.UUID, kind, raw, previous string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if identPathOK(providerID, previous) {
			return previous, nil
		}
		return "", nil
	}
	if !strings.Contains(raw, ";base64,") {
		if identPathOK(providerID, raw) {
			return raw, nil
		}
		if identPathOK(providerID, previous) {
			return previous, nil
		}
		return "", apperror.BadRequest("Rasm formati noto‘g‘ri")
	}
	ext, data, err := decodeIdentPayload(raw, true)
	if err != nil {
		return "", err
	}
	return writeIdentFile(providerID, kind+ext, data)
}

func saveIdentDocuments(providerID uuid.UUID, inputs []IdentDocumentInput, previous []IdentDocument) ([]IdentDocument, error) {
	if len(inputs) == 0 {
		return nil, apperror.BadRequest("Kamida bitta diplom yoki sertifikat PDF yuklang")
	}
	if len(inputs) > maxIdentDocs {
		return nil, apperror.BadRequest("Hujjatlar soni 5 tadan oshmasin")
	}
	prevByID := map[string]IdentDocument{}
	for _, doc := range previous {
		prevByID[doc.ID] = doc
	}
	out := make([]IdentDocument, 0, len(inputs))
	hasProof := false
	for _, input := range inputs {
		kind := strings.TrimSpace(input.Kind)
		switch kind {
		case "diploma", "certificate", "license", "other":
		default:
			return nil, apperror.BadRequest("Hujjat turi noto‘g‘ri")
		}
		if kind == "diploma" || kind == "certificate" {
			hasProof = true
		}
		title := compactSpace(input.Title)
		if n := utf8.RuneCountInString(title); n < 2 || n > 120 {
			return nil, apperror.BadRequest("Hujjat nomi 2-120 belgi oralig‘ida bo‘lishi kerak")
		}
		doc := IdentDocument{Kind: kind, Title: title}
		content := strings.TrimSpace(input.Content)
		if content != "" {
			if !strings.Contains(content, ";base64,") {
				return nil, apperror.BadRequest("PDF formati noto‘g‘ri")
			}
			ext, data, err := decodeIdentPayload(content, false)
			if err != nil {
				return nil, err
			}
			id := uuid.New()
			path, err := writeIdentFile(providerID, id.String()+ext, data)
			if err != nil {
				return nil, err
			}
			doc.ID = id.String()
			doc.Path = path
			doc.FileName = sanitizeFileName(input.FileName, "hujjat.pdf")
		} else {
			prev, ok := prevByID[strings.TrimSpace(input.Path)]
			if !ok {
				for _, item := range previous {
					if item.Path == strings.TrimSpace(input.Path) || item.ID == strings.TrimSpace(input.Path) {
						prev = item
						ok = true
						break
					}
				}
			}
			if !ok || !identPathOK(providerID, prev.Path) {
				if identPathOK(providerID, input.Path) {
					doc.ID = uuid.New().String()
					doc.Path = strings.TrimSpace(input.Path)
					doc.FileName = sanitizeFileName(input.FileName, filepath.Base(input.Path))
				} else {
					return nil, apperror.BadRequest("Hujjat fayli yuklanishi shart")
				}
			} else {
				doc.ID = prev.ID
				doc.Path = prev.Path
				doc.FileName = prev.FileName
				if name := sanitizeFileName(input.FileName, ""); name != "" {
					doc.FileName = name
				}
			}
		}
		out = append(out, doc)
	}
	if !hasProof {
		return nil, apperror.BadRequest("Kamida bitta diplom yoki sertifikat yuklang")
	}
	return out, nil
}

func decodeIdentPayload(raw string, image bool) (string, []byte, error) {
	mime, payload, ok := strings.Cut(raw, ";base64,")
	if !ok {
		return "", nil, apperror.BadRequest("Fayl formati noto‘g‘ri")
	}
	mime = strings.ToLower(mime)
	ext := ""
	max := maxIdentPDFBytes
	if image {
		max = maxIdentImageBytes
		switch {
		case strings.Contains(mime, "image/jpeg"), strings.Contains(mime, "image/jpg"):
			ext = ".jpg"
		case strings.Contains(mime, "image/webp"):
			ext = ".webp"
		case strings.Contains(mime, "image/png"):
			ext = ".png"
		default:
			return "", nil, apperror.BadRequest("Rasm faqat PNG, JPG yoki WEBP bo‘lishi mumkin")
		}
	} else {
		if !strings.Contains(mime, "application/pdf") && !strings.Contains(mime, "pdf") {
			return "", nil, apperror.BadRequest("Hujjat faqat PDF bo‘lishi mumkin")
		}
		ext = ".pdf"
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", nil, apperror.BadRequest("Faylni o‘qib bo‘lmadi")
	}
	if len(data) == 0 {
		return "", nil, apperror.BadRequest("Fayl bo‘sh")
	}
	if len(data) > max {
		if image {
			return "", nil, apperror.BadRequest("Rasm 10 MB dan oshmasin")
		}
		return "", nil, apperror.BadRequest("PDF 5 MB dan oshmasin")
	}
	return ext, data, nil
}

func writeIdentFile(providerID uuid.UUID, name string, data []byte) (string, error) {
	dir := identUploadsDir(providerID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", apperror.Internal()
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		return "", apperror.Internal()
	}
	return identPrefix(providerID) + name, nil
}

func cleanupIdentFiles(oldItem, newItem Identification) {
	keep := map[string]struct{}{
		newItem.PassportImage: {},
		newItem.SelfieImage:   {},
	}
	for _, doc := range newItem.Documents {
		keep[doc.Path] = struct{}{}
	}
	removeIdentFile(oldItem.ProviderID, oldItem.PassportImage, keep)
	removeIdentFile(oldItem.ProviderID, oldItem.SelfieImage, keep)
	for _, doc := range oldItem.Documents {
		removeIdentFile(oldItem.ProviderID, doc.Path, keep)
	}
}

func removeIdentFile(providerID uuid.UUID, path string, keep map[string]struct{}) {
	if _, ok := keep[path]; ok {
		return
	}
	if !identPathOK(providerID, path) {
		return
	}
	name := filepath.Base(path)
	_ = os.Remove(filepath.Join(identUploadsDir(providerID), name))
}

func identUploadsDir(providerID uuid.UUID) string {
	rel := filepath.Join("uploads", "service-provider-ident", providerID.String())
	candidates := []string{
		rel,
		filepath.Join("1. Backend", rel),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append([]string{filepath.Join(wd, rel)}, candidates...)
	}
	for _, dir := range candidates {
		parent := filepath.Dir(filepath.Dir(dir))
		if _, err := os.Stat(parent); err == nil {
			return dir
		}
	}
	return rel
}

func sanitizeFileName(raw, fallback string) string {
	name := filepath.Base(strings.TrimSpace(raw))
	name = strings.ReplaceAll(name, "..", "")
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return fallback
	}
	if utf8.RuneCountInString(name) > 120 {
		runes := []rune(name)
		name = string(runes[:120])
	}
	return name
}
