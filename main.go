package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/cors"
	"github.com/go-sql-driver/mysql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func CreateStore(db *sql.DB) *Store {
	return &Store{
		NewUserStore(db),
	}
}

type UserStore struct {
	*sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db,
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
	GetUserByUsername(username string) (UserItem, error)
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

	connStr := mysql.Config{
		User:                 "u6ncknqjamhqpa3d",
		Passwd:               "O1Bo5YwBLl31ua5agKoq",
		Net:                  "tcp",
		Addr:                 "bnouoawh6epgx2ipx4hl-mysql.services.clever-cloud.com:3306",
		DBName:               "bnouoawh6epgx2ipx4hl", // equivalent to chatbdd
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", connStr.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the PostgreSQL database!")

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

	// Add a new user

	http.ListenAndServe(":2345", handler)
}

// DATABASE OPERATION
func (t *UserStore) AddUser(item UserItem) (int, error) {
	res, err := t.DB.Exec("INSERT INTO Users (username, password) VALUES (?, ?)", item.Username, item.Password)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (t *UserStore) GetUserByUsername(username string) (UserItem, error) {
	var user UserItem

	err := t.QueryRow("SELECT id, username, password FROM Users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Password)

	if err == sql.ErrNoRows {
		// User not found
		return UserItem{}, nil
	} else if err != nil {
		// Handle other database errors
		return UserItem{}, err
	}

	return user, nil
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Extract registration data
	username := r.FormValue("username")
	password := r.FormValue("password")
	existentUser, err := h.Store.GetUserByUsername(username)
	if err != nil {

		h.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{"message": "L'utilisateur existe déjà", "code": http.StatusBadRequest})
		return
	} else if existentUser.Username != username {
		userID, err := h.Store.AddUser(UserItem{Username: username, Password: password})
		if err != nil {
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{"message": "l'utilisateur n'a pu être ajouté"})
			return
		}
		// Respond with a success message
		h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Registration successful", "userID": userID})
	}
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
