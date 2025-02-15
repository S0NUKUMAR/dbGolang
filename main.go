package main


import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)



type User struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

func main() {
	log.Println("Starting application...")

	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run migrations
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"database.sql", // Path to migration files
		"postgres", driver,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Apply migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	log.Println("Migrations applied successfully")

	// Set up routes
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", httpMiddleware(router)))
}

// Rest of your code...


func httpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter , r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w,r)
	})
}

func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter , r *http.Request){
		rows , err := db.Query("SELECT * FROM users")
		if err!=nil {
			//TODO : fix this problem 
			log.Fatal(err)
		}
		defer rows.Close()

		users:= []User{}
		for rows.Next() {
			var u User 
			if err:= rows.Scan(&u.Id, &u.Name , &u.Email); err!=nil{
				log.Fatal(err)
			}
			users = append(users, u)
		}
		if err:= rows.Err(); err!=nil{
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(users)
	}
}



func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		json.NewDecoder(r.Body).Decode(&user)
		
		var id int
		err:= db.QueryRow("INSERT INTO users (name , email) VALUES ($1, $2) RETURNING id", user.Name, user.Email).Scan(&id)
		if err!=nil {
			log.Fatal(err)
		}
		user.Id = id
		json.NewEncoder(w).Encode(user)
	}
}
