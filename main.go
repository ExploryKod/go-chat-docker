package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func CreateStore(db *sql.DB) *Store {
	return &Store{
		NewUserStore(db),
	}
}

type UserItem struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Admin    *int   `json:"admin"`
}

type UserStoreInterface interface {
	AddUser(item UserItem) (int, error)
	GetUsers() ([]UserItem, error)
	GetUserByUsername(username string) (UserItem, error)
	DeleteUserById(id int) error
	UpdateUser(item UserItem) error
}

type Store struct {
	UserStoreInterface
}

type Handler struct {
	*chi.Mux
	*Store
}

func main() {
	// Replace these connection parameters with your actual PostgreSQL credentials
	//connStr := "postgres://database_postgre_pn0t_user:0bITMm4I5lLLfFhfvCYkHQtJNcxzHYX3@dpg-ckor2m41tcps73e73qh0-a/database_postgre_pn0t?sslmode=disable"

	connBDD := mysql.Config{
		User:                 "u6ncknqjamhqpa3d",
		Passwd:               "O1Bo5YwBLl31ua5agKoq",
		Net:                  "tcp",
		Addr:                 "bnouoawh6epgx2ipx4hl-mysql.services.clever-cloud.com:3306",
		DBName:               "bnouoawh6epgx2ipx4hl", // equivalent to chatbdd
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", connBDD.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the MySQL database!")

	store := CreateStore(db)

	// ROUTES
	handler := &Handler{
		chi.NewRouter(),
		store,
	}

	handler.Use(middleware.Logger)
	handler.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // initialement en false
		MaxAge:           300,  // Maximum value not ignored by any of major browsers
	}))
	handler.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello My name is Golang"))
	})
	handler.Post("/auth/register", handler.RegisterHandler)
	handler.Post("/auth/logged", handler.LoginHandler())

	http.ListenAndServe(":2345", handler)
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
