package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "test_db"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to connect to test database: %v", err)
		} else {
			fmt.Printf("Failed to connect to test database: %v\n", err)
			os.Exit(1)
		}
	}

	// Clean up existing data
	db.Exec("DROP TABLE IF EXISTS users")
	db.Exec("DROP TABLE IF EXISTS products")
	err = db.AutoMigrate(&User{}, &Product{})
	if err != nil {
		if t != nil {
			t.Fatalf("Failed to migrate test database: %v", err)
		} else {
			fmt.Printf("Failed to migrate test database: %v\n", err)
			os.Exit(1)
		}
	}
}

func TestLogin(t *testing.T) {
	router := setupRouter()
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/products", getProducts).Methods("GET")
	router.HandleFunc("/products", createProduct).Methods("POST")
	router.HandleFunc("/products/{id}", updateProduct).Methods("PUT")
	router.HandleFunc("/products/{id}", deleteProduct).Methods("DELETE")

	request, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", response.Code)
	}

	fmt.Println(response.Body.String())
	cleanupTest()
}

func cleanupTest() {
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/products", getProducts).Methods("GET")
	router.HandleFunc("/products", createProduct).Methods("POST")
	router.HandleFunc("/products/{id}", updateProduct).Methods("PUT")
	router.HandleFunc("/products/{id}", deleteProduct).Methods("DELETE")
	// Add product routes here if you implement them
	return router
}

func TestMain(m *testing.M) {
	setupTestDB(nil)
	code := m.Run()
	cleanupTest()
	os.Exit(code)
}
