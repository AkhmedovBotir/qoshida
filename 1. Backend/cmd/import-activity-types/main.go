package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/console"
	"qoshida/backend/internal/database"
	"qoshida/backend/internal/modules/activitytype"
)

func main() {
	log := console.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fail("config", err)
		os.Exit(1)
	}

	log.Banner("QOSHIDA IMPORT FAOLIYAT TURLARI", cfg.AppEnv, ":"+cfg.Port, cfg.DBHost, cfg.DBPort, cfg.DBName)

	path, err := activitytype.ResolveJSONPath(os.Getenv("ACTIVITY_TYPES_JSON"))
	if err != nil {
		log.Fail("json", err)
		os.Exit(1)
	}
	log.Info("json", path)

	items, err := activitytype.LoadJSONFile(path)
	if err != nil {
		log.Fail("json", err)
		os.Exit(1)
	}
	log.OK("json", fmt.Sprintf("%d ta yozuv o'qildi", len(items)))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fail("database", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.OK("database", cfg.DBName+" ulandi")

	report, err := database.Migrate(ctx, pool)
	if err != nil {
		log.Fail("migrate", err)
		os.Exit(1)
	}
	for _, name := range report.Applied {
		log.OK("migrate", name)
	}
	for _, name := range report.Skipped {
		log.Skip("migrate", name)
	}

	service := activitytype.NewService(activitytype.NewRepository(pool))
	result, err := service.Import(ctx, items)
	if err != nil {
		log.Fail("import", err)
		os.Exit(1)
	}

	log.OK("import", fmt.Sprintf("qo'shildi: %d · yangilandi: %d · jami: %d", result.Inserted, result.Updated, result.Total))
}
