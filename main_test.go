package main

import (
	"bytes"
	"encoding/json"
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
	// Use environment variables with fallbacks
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db" // default to service name in docker-compose
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
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up existing data
	db.Exec("DROP TABLE IF EXISTS users")
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
}

func cleanupTest(t *testing.T) {
	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get database instance: %v", err)
		return
	}
	err = sqlDB.Close()
	if err != nil {
		t.Errorf("Failed to close database connection: %v", err)
	}
}

func setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers).Methods("GET")
	router.HandleFunc("/users", createUser).Methods("POST")
	router.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	return router
}

func TestCreateAndGetUser(t *testing.T) {
	setupTestDB(t)
	defer cleanupTest(t)

	router := setupRouter()

	// Create user
	user := User{Name: "Test User", Email: "test@example.com"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var createdUser User
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if createdUser.ID == 0 {
		t.Fatalf("expected user ID to be set")
	}

	// Get users
	req = httptest.NewRequest("GET", "/users", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(users) != 1 || users[0].Name != "Test User" {
		t.Fatalf("expected to find created user, got %+v", users)
	}
}

func TestUpdateAndDeleteUser(t *testing.T) {
	setupTestDB(t)
	defer cleanupTest(t)

	router := setupRouter()

	// Create user first
	user := User{Name: "ToUpdate", Email: "update@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Update user
	updated := User{Name: "Updated", Email: "updated@example.com"}
	body, _ := json.Marshal(updated)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/users/%d", user.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var updatedUser User
	if err := json.NewDecoder(resp.Body).Decode(&updatedUser); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if updatedUser.Name != "Updated" {
		t.Fatalf("expected name to be Updated, got %s", updatedUser.Name)
	}

	// Delete user
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/users/%d", user.ID), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["message"] != "User deleted" {
		t.Fatalf("expected delete confirmation, got %v", result)
	}
}

func TestMain(m *testing.M) {
	// Setup before any tests run
	setupTestDB(nil)

	// Run all tests
	code := m.Run()

	// Cleanup after all tests finish
	cleanupTest(nil)

	// Exit with the test status code
	os.Exit(code)
}
