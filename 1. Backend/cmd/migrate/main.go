package main

import (
	"context"
	"os"
	"time"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/console"
	"qoshida/backend/internal/database"
)

func main() {
	log := console.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fail("config", err)
		os.Exit(1)
	}

	log.Banner("QOSHIDA MIGRATE", cfg.AppEnv, ":"+cfg.Port, cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Info("migrate", "mavjud jadvallar va qatorlar o'chirilmaydi")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		log.OK("applied", name)
	}
	for _, name := range report.Skipped {
		log.Skip("skip", name+" (allaqachon qo'llangan)")
	}

	if len(report.Applied) == 0 {
		log.OK("migrate", "yangi o'zgarish yo'q, baza saqlanib qoldi")
	} else {
		log.OK("migrate", "tayyor, ma'lumotlar o'chirilmadi")
	}
}
