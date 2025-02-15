package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

type User struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

func main(){
	log.Println("db based example")
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err!= nil{
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	if err!=nil{
		log.Fatal(err)
	}

	log.Println("DB created")
	router:= mux.NewRouter()

	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")
	
	log.Fatal(http.ListenAndServe(":8000", httpMiddleware(router)))
}


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
