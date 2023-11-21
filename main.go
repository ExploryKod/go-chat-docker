package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
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

//type UserItem struct {
//	ID       int    `json:"id"`
//	Username string `json:"username"`
//	Password string `json:"password"`
//	Admin    *int   `json:"admin"`
//}

type UserItem struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Admin    int    `json:"admin"`
	Email    string `json:"email"`
}

type UserStoreInterface interface {
	// Users
	AddUser(item UserItem) (int, error)
	GetUsers() ([]UserItem, error)
	GetUserByUsername(username string) (UserItem, error)
	DeleteUserById(id int) error
	UpdateUser(item UserItem) error
	// Rooms
	AddRoom(item RoomItem) (int, error)
	GetRoomByName(name string) (RoomItem, error)
	GetRoomById(id int) (RoomItem, error)
	DeleteRoomById(id int) error
	AddUserToRoom(roomID int, userID int) error
	GetUsersFromRoom(roomID int) ([]UserItem, error)
	GetOneUserFromRoom(roomID int, userID int) (UserItem, error)
	GetRooms() ([]RoomItem, error)
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

	// Initiate Ws
	wsServer := NewWebsocketServer()
	go wsServer.Run()

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

	handler.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))

		r.Use(jwtauth.Authenticator)

		r.Get("/user-list", handler.GetUsers())
		r.Delete("/delete-user/{id}", handler.DeleteUserHandler())
		r.Get("/update-user", handler.UpdateHandler)

		r.Get("/chat/{id}", handler.JoinRoomHandler())
		r.Get("/chat/rooms", handler.GetRooms())
		r.Post("/chat/create", handler.CreateRoomHandler())
		r.Delete("/delete-room/{id}", handler.DeleteRoomHandler())
	})
	handler.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsServer, w, r)
	})

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
