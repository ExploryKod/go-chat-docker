package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Extract registration data
	username := r.FormValue("username")
	password := r.FormValue("password")
	existentUser, err := h.Store.GetUserByUsername(username)
	if err != nil {
		//http.Error(w, "User already exists", http.StatusBadRequest)
		//errorResponse := ErrorResponse{
		//	Message: "L'utilisateur ${username} existe déjà",
		//	Code:    http.StatusBadRequest,
		//}
		//h.jsonResponse(w, http.StatusBadRequest, errorResponse)

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

var tokenAuth *jwtauth.JWTAuth

const Secret = "mysecretamaury"

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(Secret), nil)
}

func MakeToken(name string) string {
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"username": name})
	return tokenString
}

func (h *Handler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract username and password from the request body or form data
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Validate user credentials against the database
		user, err := h.Store.GetUserByUsername(username)
		if err != nil {
			// Handle database error
			h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"message": "Internal Server Error" + err.Error(),
			})
			return
		}

		if user.Username == "" || user.Password == "" {
			http.Error(w, "Il reste des champs vide", http.StatusBadRequest)
			return
		}

		// Check if the user exists and the password matches
		if user.Username == username && user.Password == password {
			token := MakeToken(username)

			http.SetCookie(w, &http.Cookie{
				HttpOnly: true,
				Expires:  time.Now().Add(7 * 24 * time.Hour),
				SameSite: http.SameSiteLaxMode,
				// Uncomment below for HTTPS:
				// Secure: true,
				Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
				Value: token,
			})
			// Successful login

			// Convert role (admin column) to string
			var roleStr string

			if user.Admin != nil {
				roleStr = strconv.Itoa(*user.Admin)
			} else {
				roleStr = "0"
			}

			response := map[string]string{"message": "Vous êtes bien connecté", "redirect": "/", "token": token, "role": roleStr}
			h.jsonResponse(w, http.StatusOK, response)
		} else if user.Password != password {
			// Failed login
			h.jsonResponse(w, http.StatusUnauthorized, map[string]interface{}{
				"message": "Mot de passe incorrect",
			})
		} else if user.Username != username {
			// Failed login
			h.jsonResponse(w, http.StatusUnauthorized, map[string]interface{}{
				"message": "Nom d'utilisateur incorrect",
			})
		} else {
			// Failed login
			h.jsonResponse(w, http.StatusUnauthorized, map[string]interface{}{
				"message": "Nom d'utilisateur et mot de passe incorrects",
			})
		}
	}
}

func (h *Handler) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.Store.GetUsers()
		if err != nil {
			// Handle database error
			h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"message": "Internal Server Error",
			})
			return
		}

		h.jsonResponse(w, http.StatusOK, users)
	}
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	// Extract registration data
	username := r.FormValue("username")
	password := r.FormValue("password")
	userID := r.FormValue("id")
	id, _ := strconv.Atoi(userID)
	existentUser, _ := h.Store.GetUserByUsername(username)
	if existentUser.Username != "" {

		err := h.Store.UpdateUser(UserItem{ID: id, Username: username, Password: password})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Respond with a success message
		h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Update successful"})
	} else {
		http.Error(w, "No user with this id found", http.StatusBadRequest)
		return
	}
}

func (h *Handler) DeleteUserHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		QueryId := chi.URLParam(request, "id")
		//QueryId := request.FormValue("id")
		id, _ := strconv.Atoi(QueryId)

		err := h.Store.DeleteUserById(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		h.jsonResponse(writer, http.StatusOK, map[string]interface{}{"message": "User deleted"})
		http.Redirect(writer, request, "/user-list", http.StatusSeeOther)

	}
}
