package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/mail"
	"strconv"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Extract registration data
	username := r.FormValue("username")
	password := r.FormValue("password")
	emailFromClient := r.FormValue("email")
	emailChecked := valid(emailFromClient)
	var email string
	if emailChecked == false || emailFromClient == "" {
		email = "waiting@noemail.com"
	} else {
		email = emailFromClient
	}
	existentUser, err := h.Store.GetUserByUsername(username)
	if err != nil {
		h.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{"message": "L'utilisateur existe déjà", "code": http.StatusBadRequest})
		return
	} else if existentUser.Username != username {
		userID, err := h.Store.AddUser(UserItem{Username: username, Password: password, Email: email, Admin: 0})
		if err != nil {
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{"message": "l'utilisateur n'a pu être ajouté"})
			return
		}
		// Respond with a success message
		h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Vous êtes bien inscris: connectez-vous pour chatter", "userID": userID})
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
			roleStr := strconv.Itoa(user.Admin)
			email := user.Email

			response := map[string]string{"message": "Vous êtes bien connecté", "redirect": "/", "token": token, "admin": roleStr, "email": email}
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

func (h *Handler) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		username := r.FormValue("username")
		role := r.FormValue("admin")
		userID := r.FormValue("id")
		email := r.FormValue("email")
		id, _ := strconv.Atoi(userID)
		admin, _ := strconv.Atoi(role)

		err := h.Store.UpdateUser(UserItem{ID: id, Username: username, Admin: admin, Email: email})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Utilisateur modifié", "username": username, "statut": role, "email": email})

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
