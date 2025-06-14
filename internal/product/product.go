package product

import (
	"encoding/json"
	"net/http"

	"sksingh/dbExample/internal/db"
	"sksingh/dbExample/internal/middleware"

	"github.com/gorilla/mux"
)

type ProductInterface interface {
	GetProducts() ([]Product, error)
	CreateProduct(product Product) (Product, error)
	UpdateProduct(product Product) (Product, error)
	DeleteProduct(product Product) (Product, error)
}

type ProductService struct {
	*db.DBService
}

type Product struct {
	ID    uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name  string `json:"name"`
	Price uint   `json:"price"`
}

func NewProductService(db *db.DBService) *ProductService {
	return &ProductService{DBService: db}
}

func (s *ProductService) RegisterProductRoutes(router *mux.Router) {
	router.HandleFunc("/products", middleware.JWTMiddleware(s.getProducts)).Methods("GET")
	router.HandleFunc("/products", middleware.JWTMiddleware(s.createProduct)).Methods("POST")
	router.HandleFunc("/products/{id}", middleware.JWTMiddleware(s.updateProduct)).Methods("PUT")
	router.HandleFunc("/products/{id}", middleware.JWTMiddleware(s.deleteProduct)).Methods("DELETE")
}

func (s *ProductService) getProducts(w http.ResponseWriter, r *http.Request) {
	var products []Product
	if err := s.Db.Find(&products).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(products)
}

func (s *ProductService) createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	//strictly check the request payload
	if product.Name == "" || product.Price == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	//create the product
	if err := s.Db.Create(&product).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(product)
}

func (s *ProductService) updateProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var product Product
	if err := s.Db.First(&product, params["id"]).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}
	var updatedProduct Product
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	product.Name = updatedProduct.Name
	product.Price = updatedProduct.Price
	if err := s.Db.Save(&product).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(product)
}

func (s *ProductService) deleteProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var product Product
	if err := s.Db.First(&product, params["id"]).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}
	if err := s.Db.Delete(&product).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted"})
}
