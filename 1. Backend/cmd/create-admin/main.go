package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/console"
	"qoshida/backend/internal/database"
	"qoshida/backend/internal/modules/admin"
)

func main() {
	log := console.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fail("config", err)
		os.Exit(1)
	}
	if strings.TrimSpace(cfg.AdminPassword) == "" {
		log.Fail("ADMIN_PASSWORD", fmt.Errorf(".env da ADMIN_PASSWORD majburiy"))
		os.Exit(1)
	}

	log.Banner("QOSHIDA CREATE-ADMIN", cfg.AppEnv, ":"+cfg.Port, cfg.DBHost, cfg.DBPort, cfg.DBName)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
	for _, name := range report.Skipped {
		log.Skip("migrate", name)
	}
	for _, name := range report.Applied {
		log.OK("migrate", name)
	}

	service := admin.NewService(admin.NewRepository(pool))
	created, err := service.CreateGeneral(ctx, admin.CreateRequest{
		FirstName: cfg.AdminFirstName,
		LastName:  cfg.AdminLastName,
		Phone:     cfg.AdminPhone,
		Username:  cfg.AdminUsername,
		Password:  cfg.AdminPassword,
	})
	if err != nil {
		log.Fail("create-admin", err)
		os.Exit(1)
	}

	log.OK("create-admin", fmt.Sprintf("%s (%s) · %s", created.Username, created.Phone, created.Role))
}
