package category

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type oidRef struct {
	OID string `json:"$oid"`
}

type rawCategory struct {
	ID       oidRef  `json:"_id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Image    *string `json:"image"`
	Parent   *oidRef `json:"parent"`
	Status   string  `json:"status"`
	Censored bool    `json:"censored"`
}

func LoadJSONFile(path string) ([]Category, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("json o'qilmadi: %w", err)
	}

	var rows []rawCategory
	if err = json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("json formati noto'g'ri: %w", err)
	}

	items := make([]Category, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		slug := strings.TrimSpace(row.Slug)
		status := strings.TrimSpace(row.Status)
		if status == "" {
			status = StatusActive
		}
		if row.ID.OID == "" || name == "" {
			continue
		}
		if slug == "" {
			slug = slugify(name)
		}
		item := Category{
			SourceID: row.ID.OID,
			Name:     name,
			Slug:     slug,
			Censored: row.Censored,
			Status:   status,
		}
		if row.Parent != nil {
			item.ParentSourceID = strings.TrimSpace(row.Parent.OID)
		}
		if row.Image != nil && strings.TrimSpace(*row.Image) != "" {
			url, imgErr := saveDataImage(item.SourceID, *row.Image, 0)
			if imgErr != nil {
				return nil, fmt.Errorf("%s rasmi: %w", item.Name, imgErr)
			}
			item.ImageURL = url
		}
		items = append(items, item)
	}
	return items, nil
}

func ResolveJSONPath(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		if _, err := os.Stat(explicit); err == nil {
			return explicit, nil
		}
		return "", fmt.Errorf("fayl topilmadi: %s", explicit)
	}

	candidates := []string{
		filepath.Join("scripts", "ttsa.categories.json"),
		filepath.Join(".", "scripts", "ttsa.categories.json"),
		filepath.Join("1. Backend", "scripts", "ttsa.categories.json"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "scripts", "ttsa.categories.json"))
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("scripts/ttsa.categories.json topilmadi")
}
