package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var migrationFS embed.FS

type MigrationFile struct {
	ID   int
	Name string
	SQL  string
}

type MigrateReport struct {
	Applied []string
	Skipped []string
}

func LoadMigrations() ([]MigrationFile, error) {
	entries, err := fs.ReadDir(migrationFS, "sql")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	files := make([]MigrationFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, err := fs.ReadFile(migrationFS, "sql/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		id, name, err := parseMigrationName(entry.Name())
		if err != nil {
			return nil, err
		}
		files = append(files, MigrationFile{ID: id, Name: name, SQL: string(raw)})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].ID < files[j].ID })
	return files, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) (MigrateReport, error) {
	var report MigrateReport

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return report, fmt.Errorf("schema_migrations: %w", err)
	}
	if _, err := pool.Exec(ctx, `ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT ''`); err != nil {
		return report, fmt.Errorf("schema_migrations.name: %w", err)
	}

	applied, err := appliedIDs(ctx, pool)
	if err != nil {
		return report, err
	}

	files, err := LoadMigrations()
	if err != nil {
		return report, err
	}

	for _, file := range files {
		label := fmt.Sprintf("%06d_%s", file.ID, file.Name)
		if applied[file.ID] {
			report.Skipped = append(report.Skipped, label)
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return report, fmt.Errorf("begin %s: %w", label, err)
		}

		if _, err = tx.Exec(ctx, file.SQL); err != nil {
			_ = tx.Rollback(ctx)
			return report, fmt.Errorf("apply %s: %w", label, err)
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO schema_migrations (id, name) VALUES ($1, $2)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
		`, file.ID, file.Name); err != nil {
			_ = tx.Rollback(ctx)
			return report, fmt.Errorf("record %s: %w", label, err)
		}
		if err = tx.Commit(ctx); err != nil {
			return report, fmt.Errorf("commit %s: %w", label, err)
		}

		report.Applied = append(report.Applied, label)
	}

	return report, nil
}

func appliedIDs(ctx context.Context, pool *pgxpool.Pool) (map[int]bool, error) {
	rows, err := pool.Query(ctx, `SELECT id FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	defer rows.Close()

	out := make(map[int]bool)
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func parseMigrationName(filename string) (int, string, error) {
	base := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("migration nomi noto'g'ri: %s (000001_name.sql)", filename)
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("migration raqami noto'g'ri: %s", filename)
	}
	return id, parts[1], nil
}
