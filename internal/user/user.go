package user

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"sksingh/dbExample/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type User struct {
	ID    uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserService struct {
	*db.DBService
}

func NewUserService(db *db.DBService) *UserService {
	return &UserService{DBService: db}
}

func (s *UserService) RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/login", s.login).Methods("POST")
	router.HandleFunc("/register", s.register).Methods("POST")
	router.HandleFunc("/logout", s.logout).Methods("POST")
}

func (s *UserService) login(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	// Create a token for the user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Minute * 1).Unix(),
		"iat": time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	// Set the token in the cookie (optional)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(1 * time.Minute),
		Path:     "/",
		HttpOnly: true,
	})
	//send the login success message
	json.NewEncoder(w).Encode(map[string]string{"message": "User logged in successfully"})
}

func (s *UserService) register(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	//check if the user already exists
	if err := s.Db.Where("email = ?", user.Email).First(&user).Error; err == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User already exists"})
		return
	}
	if err := s.Db.Create(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (s *UserService) logout(w http.ResponseWriter, r *http.Request) {
	// Clear the token from the cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-time.Hour * 24),
	})
	json.NewEncoder(w).Encode(map[string]string{"message": "User logged out successfully"})
}

func (s *UserService) getUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := s.Db.Where("email = ?", user.Email).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (s *UserService) updateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := s.Db.Where("email = ?", user.Email).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}
	json.NewEncoder(w).Encode(user)
}
