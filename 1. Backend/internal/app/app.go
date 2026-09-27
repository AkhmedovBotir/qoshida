package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"qoshida/backend/internal/config"
	"qoshida/backend/internal/console"
	"qoshida/backend/internal/database"
	"qoshida/backend/internal/middleware"
	"qoshida/backend/internal/modules/activitytype"
	"qoshida/backend/internal/modules/admin"
	"qoshida/backend/internal/modules/auth"
	"qoshida/backend/internal/modules/category"
	"qoshida/backend/internal/modules/commenttemplate"
	"qoshida/backend/internal/modules/customer"
	"qoshida/backend/internal/modules/delivery"
	"qoshida/backend/internal/modules/eskiz"
	"qoshida/backend/internal/modules/health"
	"qoshida/backend/internal/modules/kontragent"
	"qoshida/backend/internal/modules/localshop"
	"qoshida/backend/internal/modules/manager"
	"qoshida/backend/internal/modules/market"
	"qoshida/backend/internal/modules/product"
	"qoshida/backend/internal/modules/providerservice"
	"qoshida/backend/internal/modules/region"
	"qoshida/backend/internal/modules/seller"
	"qoshida/backend/internal/modules/serviceprovider"
	"qoshida/backend/internal/modules/shopdirector"
	"qoshida/backend/internal/modules/shopstock"
	"qoshida/backend/internal/modules/shoptemplate"
	"qoshida/backend/internal/shared/audit"
)

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	server *http.Server
	SMS    *eskiz.Service
}

func New(ctx context.Context, cfg *config.Config, log *console.Logger) (*App, error) {
	pool, err := database.Connect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	log.OK("database", cfg.DBName+" ulandi")
	audit.Init(pool)

	report, err := database.Migrate(ctx, pool)
	if err != nil {
		pool.Close()
		return nil, err
	}
	for _, name := range report.Applied {
		log.OK("migrate", name)
	}
	for _, name := range report.Skipped {
		log.Skip("migrate", name)
	}
	if len(report.Applied) == 0 {
		log.OK("migrate", "yangi o'zgarish yo'q")
	}

	adminRepo := admin.NewRepository(pool)
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)

	tokenRepo := auth.NewTokenRepository(pool)
	authService := auth.NewService(adminRepo, tokenRepo, cfg)
	authHandler := auth.NewHandler(authService)
	healthHandler := health.New(pool)
	categoryHandler := category.NewHandler(category.NewService(category.NewRepository(pool)))
	activityRepo := activitytype.NewRepository(pool)
	activityService := activitytype.NewService(activityRepo)
	activityTypeHandler := activitytype.NewHandler(activityService)
	commentTemplateHandler := commenttemplate.NewHandler(commenttemplate.NewService(commenttemplate.NewRepository(pool)))
	smsService := eskiz.New(cfg)
	if smsService.Enabled() {
		log.OK("eskiz", "SMS moduli tayyor")
	} else {
		log.Skip("eskiz", "ESKIZ_EMAIL / ESKIZ_PASSWORD yo'q")
	}

	regionRepo := region.NewRepository(pool)
	regionService := region.NewService(regionRepo)
	regionHandler := region.NewHandler(regionService)
	customerRepo := customer.NewRepository(pool)
	customerHandler := customer.NewHandler(customer.NewService(customerRepo, regionRepo), customer.NewAuthService(customerRepo, smsService, cfg))
	marketHandler := market.NewHandler(market.NewService(market.NewRepository(pool)), regionService)
	managerRepo := manager.NewRepository(pool)
	managerService := manager.NewService(managerRepo, regionRepo)
	managerAuth := manager.NewAuthService(managerRepo, smsService, cfg)
	kontragentRepo := kontragent.NewRepository(pool)
	kontragentService := kontragent.NewService(kontragentRepo, regionRepo, activityRepo)
	kontragentAuth := kontragent.NewAuthService(kontragentRepo, smsService, cfg)
	productHandler := product.NewHandler(product.NewService(product.NewRepository(pool)))
	offeringHandler := providerservice.NewHandler(providerservice.NewService(providerservice.NewRepository(pool)))
	templateHandler := shoptemplate.NewHandler(shoptemplate.NewService(shoptemplate.NewRepository(pool)))
	stockHandler := shopstock.NewHandler(shopstock.NewService(shopstock.NewRepository(pool)))
	kontragentHandler := kontragent.NewHandler(kontragentService, kontragentAuth, activityService, productHandler, categoryHandler)
	shopRepo := localshop.NewRepository(pool)
	shopService := localshop.NewService(shopRepo, regionRepo)
	shopAuth := localshop.NewAuthService(shopRepo, smsService, cfg)
	shopHandler := localshop.NewHandler(shopService, shopAuth)
	providerRepo := serviceprovider.NewRepository(pool)
	providerService := serviceprovider.NewService(providerRepo, regionRepo, activityRepo)
	providerAuth := serviceprovider.NewAuthService(providerRepo, smsService, cfg)
	providerHandler := serviceprovider.NewHandler(providerService, providerAuth, activityService)
	providerHandler.SetServices(offeringHandler)
	sellerRepo := seller.NewRepository(pool)
	sellerService := seller.NewService(sellerRepo)
	sellerAuth := seller.NewAuthService(sellerRepo, smsService, cfg)
	sellerHandler := seller.NewHandler(sellerService, sellerAuth)
	deliveryRepo := delivery.NewRepository(pool)
	deliveryService := delivery.NewService(deliveryRepo)
	deliveryAuth := delivery.NewAuthService(deliveryRepo, smsService, cfg)
	deliveryHandler := delivery.NewHandler(deliveryService, deliveryAuth)
	shopHandler.SetSellers(sellerHandler)
	shopHandler.SetDeliveries(deliveryHandler)
	shopHandler.SetCatalog(templateHandler, stockHandler, categoryHandler)
	sellerHandler.SetCatalog(templateHandler, stockHandler)
	directorRepo := shopdirector.NewRepository(pool)
	directorService := shopdirector.NewService(directorRepo, regionRepo)
	directorAuth := shopdirector.NewAuthService(directorRepo, smsService, cfg)
	directorHandler := shopdirector.NewHandler(directorService, directorAuth, kontragentHandler, shopHandler, providerHandler, sellerHandler, deliveryHandler, productHandler, categoryHandler, templateHandler, stockHandler, offeringHandler)
	managerHandler := manager.NewHandler(managerService, managerAuth, regionRepo, directorHandler, kontragentHandler, shopHandler, providerHandler, sellerHandler, deliveryHandler, productHandler, categoryHandler, templateHandler, stockHandler, offeringHandler)

	r := chi.NewRouter()
	middleware.Register(r, cfg)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	r.Route("/api/v1", func(api chi.Router) {
		api.Mount("/health", healthHandler.Routes())
		api.Mount("/auth", authHandler.Routes(cfg.AuthRateLimitRPM))
		api.Mount("/manager-auth", managerHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/shop-director-auth", directorHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/kontragent-auth", kontragentHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/local-shop-auth", shopHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/service-provider-auth", providerHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/seller-auth", sellerHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/delivery-auth", deliveryHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/mijoz-auth", customerHandler.AuthRoutes(cfg.AuthRateLimitRPM))
		api.Mount("/public", marketHandler.PublicRoutes())
		api.Group(func(cust chi.Router) {
			cust.Use(customerHandler.RequireAuth)
			cust.Route("/mijoz", func(r chi.Router) {
				customerHandler.RegisterApp(r)
				marketHandler.RegisterApp(r)
			})
		})
		api.Group(func(protected chi.Router) {
			protected.Use(authHandler.RequireAuth)
			protected.Get("/dashboard/stats", adminHandler.Stats)
			protected.Mount("/admins", adminHandler.Routes())
			protected.Mount("/regions", regionHandler.Routes())
			protected.Mount("/categories", categoryHandler.Routes())
			protected.Mount("/activity-types", activityTypeHandler.Routes())
			protected.Mount("/comment-templates", commentTemplateHandler.Routes())
			protected.Mount("/managers", managerHandler.AdminRoutes())
			protected.Mount("/shop-directors", directorHandler.AdminRoutes())
			protected.Mount("/kontragents", kontragentHandler.AdminRoutes())
			protected.Mount("/local-shops", shopHandler.AdminRoutes())
			protected.Mount("/service-providers", providerHandler.AdminRoutes())
			protected.Mount("/provider-services", offeringHandler.AdminRoutes())
			protected.Mount("/sellers", sellerHandler.AdminRoutes())
			protected.Mount("/deliveries", deliveryHandler.AdminRoutes())
			protected.Mount("/products", productHandler.AdminRoutes())
			protected.Mount("/shop-templates", templateHandler.AdminRoutes())
			protected.Mount("/shop-products", stockHandler.AdminProductRoutes())
			protected.Mount("/shop-incomings", stockHandler.AdminIncomingRoutes())
		})
		api.Group(func(mgr chi.Router) {
			mgr.Use(managerHandler.RequireAuth)
			mgr.Mount("/manager", managerHandler.AppRoutes())
		})
		api.Group(func(dir chi.Router) {
			dir.Use(directorHandler.RequireAuth)
			dir.Mount("/shop-director", directorHandler.AppRoutes())
		})
		api.Group(func(kg chi.Router) {
			kg.Use(kontragentHandler.RequireAuth)
			kg.Mount("/kontragent", kontragentHandler.AppRoutes())
		})
		api.Group(func(shop chi.Router) {
			shop.Use(shopHandler.RequireAuth)
			shop.Mount("/local-shop", shopHandler.AppRoutes())
		})
		api.Group(func(sp chi.Router) {
			sp.Use(providerHandler.RequireAuth)
			sp.Mount("/service-provider", providerHandler.AppRoutes())
		})
		api.Group(func(sl chi.Router) {
			sl.Use(sellerHandler.RequireAuth)
			sl.Mount("/seller", sellerHandler.AppRoutes())
		})
		api.Group(func(dl chi.Router) {
			dl.Use(deliveryHandler.RequireAuth)
			dl.Mount("/delivery", deliveryHandler.AppRoutes())
		})
	})

	return &App{
		cfg:  cfg,
		pool: pool,
		SMS:  smsService,
		server: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           r,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       cfg.RequestTimeout + 2*time.Second,
			WriteTimeout:      2 * time.Minute,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 13,
		},
	}, nil
}

func (a *App) Run() error {
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	err := a.server.Shutdown(ctx)
	a.pool.Close()
	return err
}
