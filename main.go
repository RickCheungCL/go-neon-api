package main

import (
	"log"
	"os"

	"github.com/joho/godotenv" // optional; add to go.mod if you want
	"github.com/rick/go-neon-api/internal/db"
	"github.com/rick/go-neon-api/internal/http"
	"github.com/rick/go-neon-api/internal/http/handlers"
	"github.com/rick/go-neon-api/internal/models"
)

func main() {
	_ = godotenv.Load()

	db.Connect()

	// Auto-migrate all tables from your Prisma schema mapping
	if err := db.DB.AutoMigrate(
		&models.User{},
		&models.Case{},
		&models.ActivityLog{},
		&models.LightFixtureType{},
		&models.CaseFixtureCount{},
		&models.InstallationDetail{},
		&models.InstallationTag{},
		&models.InstallationDetailTag{},
		&models.OnSiteVisit{},
		&models.Product{},
		&models.OnSiteVisitRoom{},
		&models.OnSiteExistingProduct{},
		&models.OnSiteSuggestedProduct{},
		&models.OnSiteVisitPhoto{},
		&models.OnSitePhotoTag{},
		&models.OnSiteVisitPhotoTagPivot{},
		&models.OnSiteLocationTag{},
		&models.Document{},
		&models.Photo{},
		&models.QuoteCounter{},
		&models.PaybackSetting{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	h := handlers.New()
	r := http.NewRouter(h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
