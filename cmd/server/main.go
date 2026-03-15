package main

import (
	"context"
	"fmt"
	"log"
	"wallet_service/config"
	"wallet_service/internal/controller"
	"wallet_service/internal/id"
	"wallet_service/internal/repo"
	"wallet_service/internal/router"
	"wallet_service/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize repository
	repoFactory := repo.NewRepoFactory(
		cfg.Repository.SegmentCount,
		repo.DBConfig{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			DBName:   cfg.Database.DBName,
		},
	)

	repoType := repo.RepoType(cfg.Repository.Type)
	ctx := context.Background()
	repository, err := repoFactory.GetRepository(ctx, repoType)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	// Initialize dependencies
	generator := id.NewGenerator()
	walletService := service.NewWalletService(repository, generator)

	// Initialize controller
	walletController := controller.NewWalletController(walletService)

	// Setup router
	r := router.Setup(walletController)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(r.Run(addr))
}
