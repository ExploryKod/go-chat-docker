package main

import (
	"database/sql"
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type Handler struct {
	*chi.Mux
	*Store
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default to port 8000 if PORT environment variable is not set
	}

	//psql 'postgresql://ExploryKod:0PqEazdVC2RJ@ep-square-block-44724621-pooler.eu-central-1.aws.neon.tech/chatdb?sslmode=require'

	//conf := mysql.Config{
	//	User:                 "u6ncknqjamhqpa3d",
	//	Passwd:               "O1Bo5YwBLl31ua5agKoq",
	//	Net:                  "tcp",
	//	Addr:                 "bnouoawh6epgx2ipx4hl-mysql.services.clever-cloud.com:3306",
	//	DBName:               "bnouoawh6epgx2ipx4hl",
	//	AllowNativePasswords: true,
	//}

	conf := mysql.Config{
		User:                 "root",
		Passwd:               os.Getenv("MARIADB_ROOT_PASSWORD"),
		Net:                  "tcp",
		Addr:                 "database:3306",
		DBName:               os.Getenv("MARIADB_DATABASE"),
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	store := CreateStore(db)
	//mux := NewHandler(store)

	type Todo struct {
		Title string
		Done  bool
	}

	type TodoPageData struct {
		PageTitle string
		Todos     []Todo
	}

	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	data := TodoPageData{
	//		PageTitle: "My TODO list",
	//		Todos: []Todo{
	//			{Title: "Task 1", Done: false},
	//			{Title: "Task 2", Done: true},
	//			{Title: "Task 3", Done: true},
	//		},
	//	}
	//	err := tmpl.Execute(w, data)
	//	if err != nil {
	//		return
	//	}
	//})
	//http.ListenAndServe(":80", nil)

	handler := &Handler{
		chi.NewRouter(),
		store,
	}

	flag.Parse()
	wsServer := NewWebsocketServer()
	go wsServer.Run()

	//r := chi.NewRouter()
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

	handler.Post("/auth/register", handler.RegisterHandler)
	handler.Post("/auth/logged", handler.LoginHandler())

	handler.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))

		r.Use(jwtauth.Authenticator)
		// use JoinHub method to join a hub
		r.Get("/chat/{id}", handler.JoinRoomHandler())
		r.Get("/chat/rooms", handler.GetRooms())
		r.Post("/chat/create", handler.CreateRoomHandler())
		r.Get("/user-list", handler.GetUsers())
		r.Delete("/delete-user/{id}", handler.DeleteUserHandler())
		r.Delete("/delete-room/{id}", handler.DeleteRoomHandler())
		r.Post("/update-user", handler.UpdateHandler())
		r.Post("/update-room", handler.UpdateRoomHandler())
		r.Post("/send-message", handler.CreateMessageHandler)
		r.Get("/chat/messages/{id}", handler.GetMessageHandler)
		r.Get("/messages/room/delete-history/{id}", handler.DeleteMessageFromRoomHandler())
	})

	handler.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsServer, w, r)
	})

	tmpl := template.Must(template.ParseFiles("./layout.html"))
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TodoPageData{
			PageTitle: "My TODO list",
			Todos: []Todo{
				{Title: "Task 1", Done: false},
				{Title: "Task 2", Done: true},
				{Title: "Task 3", Done: true},
			},
		}
		err := tmpl.Execute(w, data)
		if err != nil {
			return
		}
	})

	server := &http.Server{
		Addr:              ":8000", // Replace with your desired address
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           handler, // Use the chi router as the handler
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
