package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"smart-pet-monitoring/backend/internal/bootstrap"
	"smart-pet-monitoring/backend/internal/config"
	httpdelivery "smart-pet-monitoring/backend/internal/delivery/http"
	"smart-pet-monitoring/backend/internal/repository/postgres"
	"smart-pet-monitoring/backend/internal/security"
	"smart-pet-monitoring/backend/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.Ping(pingCtx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(cfg.UploadDir, "pets"), 0o755); err != nil {
		log.Fatalf("create upload directory: %v", err)
	}

	ownerRepo := postgres.NewOwnerRepository(db)
	feedRepo := postgres.NewFeedRepository(db)
	commandRepo := postgres.NewCommandRepository(db)
	dashboardRepo := postgres.NewDashboardRepository(db)
	healthRepo := postgres.NewHealthRepository(db)
	profileRepo := postgres.NewProfileRepository(db)

	if cfg.SeedDemoOwner {
		seedCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		owner, err := bootstrap.EnsureDemoOwner(seedCtx, ownerRepo, cfg.DemoOwnerName, cfg.DemoOwnerEmail, cfg.DemoOwnerPassword)
		if err != nil {
			cancel()
			log.Fatalf("seed demo owner: %v", err)
		}
		if err := bootstrap.EnsureDemoProfile(seedCtx, db, owner, cfg.DeviceID); err != nil {
			cancel()
			log.Fatalf("seed demo profile: %v", err)
		}
		cancel()
	}

	jwtService := security.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresIn)
	authUsecase := usecase.NewAuthUsecase(ownerRepo, jwtService)
	feedUsecase := usecase.NewFeedUsecase(feedRepo, cfg.HistoryDefaultLimit, cfg.HistoryMaxLimit)
	controlUsecase := usecase.NewControlUsecase(commandRepo)
	dashboardUsecase := usecase.NewDashboardUsecase(dashboardRepo, cfg.DeviceID)
	healthUsecase := usecase.NewHealthUsecase(healthRepo)
	profileUsecase := usecase.NewProfileUsecase(profileRepo, cfg.DeviceID)

	gin.SetMode(cfg.GinMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	handler := httpdelivery.NewHandler(authUsecase, feedUsecase, controlUsecase, dashboardUsecase, healthUsecase, profileUsecase, jwtService, cfg)
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("backend listening on http://localhost:%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown server: %v", err)
	}
}
