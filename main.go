package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello My name is Golang"))
	})

	// Replace these connection parameters with your actual PostgreSQL credentials
	connStr := "postgres://database_postgre_pn0t_user:0bITMm4I5lLLfFhfvCYkHQtJNcxzHYX3@dpg-ckor2m41tcps73e73qh0-a/database_postgre_pn0t?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the PostgreSQL database!")

	// Perform your database operations here...

	http.ListenAndServe(":2345", r)
}
