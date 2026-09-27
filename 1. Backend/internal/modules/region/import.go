package region

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

type rawRegion struct {
	ID     oidRef  `json:"_id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Parent *oidRef `json:"parent"`
	Code   string  `json:"code"`
	Status string  `json:"status"`
}

func LoadJSONFile(path string) ([]Region, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("json o'qilmadi: %w", err)
	}

	var rows []rawRegion
	if err = json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("json formati noto'g'ri: %w", err)
	}

	items := make([]Region, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		code := strings.TrimSpace(row.Code)
		status := strings.TrimSpace(row.Status)
		if status == "" {
			status = StatusActive
		}
		if row.ID.OID == "" || name == "" || code == "" {
			continue
		}
		item := Region{
			SourceID: row.ID.OID,
			Name:     name,
			Type:     strings.TrimSpace(row.Type),
			Code:     code,
			Status:   status,
		}
		if row.Parent != nil {
			item.ParentSourceID = strings.TrimSpace(row.Parent.OID)
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
		filepath.Join("scripts", "ttsa.regions.json"),
		filepath.Join(".", "scripts", "ttsa.regions.json"),
		filepath.Join("1. Backend", "scripts", "ttsa.regions.json"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "scripts", "ttsa.regions.json"))
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("scripts/ttsa.regions.json topilmadi")
}
