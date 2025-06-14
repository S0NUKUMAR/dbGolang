package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"sksingh/dbExample/internal/db"
	"sksingh/dbExample/internal/middleware"
	"sksingh/dbExample/internal/product"
	"sksingh/dbExample/internal/user"

	"github.com/gorilla/mux"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	JWTSecret  string
	ServerPort string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBPort:     getEnv("DB_PORT", "5432"),
		JWTSecret:  getEnv("JWT_SECRET", "your_secret_key_here"),
		ServerPort: getEnv("SERVER_PORT", "8000"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	cfg := LoadConfig()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Kolkata",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	dbService, err := db.NewDBService(dsn)
	if err != nil {
		log.Fatalf("failed to get db service: %v", err)
	}

	err = dbService.Connect(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate the schema
	err = dbService.AutoMigrate(&user.User{}, &product.Product{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	router := mux.NewRouter()
	userService := user.NewUserService(dbService)
	userService.RegisterAuthRoutes(router)

	productService := product.NewProductService(dbService)
	productService.RegisterProductRoutes(router)

	log.Printf("Server started on :%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, middleware.HttpMiddleware(router)))
}
