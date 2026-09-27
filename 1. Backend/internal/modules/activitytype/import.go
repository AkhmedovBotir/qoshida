package activitytype

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

type rawActivityType struct {
	ID     oidRef `json:"_id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Status string `json:"status"`
}

func LoadJSONFile(path string) ([]ActivityType, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("json o'qilmadi: %w", err)
	}

	var rows []rawActivityType
	if err = json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("json formati noto'g'ri: %w", err)
	}

	items := make([]ActivityType, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		icon := strings.TrimSpace(row.Icon)
		status := strings.TrimSpace(row.Status)
		if status == "" {
			status = StatusActive
		}
		if row.ID.OID == "" || name == "" {
			continue
		}
		items = append(items, ActivityType{
			SourceID: row.ID.OID,
			Name:     name,
			Icon:     icon,
			Status:   status,
		})
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
		filepath.Join("scripts", "ttsa.contragenttypes.json"),
		filepath.Join(".", "scripts", "ttsa.contragenttypes.json"),
		filepath.Join("1. Backend", "scripts", "ttsa.contragenttypes.json"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "scripts", "ttsa.contragenttypes.json"))
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("scripts/ttsa.contragenttypes.json topilmadi")
}
