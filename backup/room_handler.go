package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
	"strconv"
)

func (h *Handler) JoinRoomHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "id")
		var id, err = strconv.Atoi(roomID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		room, err := h.Store.GetRoomById(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, claims, _ := jwtauth.FromContext(r.Context())
		if username, ok := claims["username"].(string); ok {
			user, err := h.Store.GetUserByUsername(username)
			if err != nil {
				// Handle database error
				h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
					"message": "Internal Server Error",
				})
				return
			}
			fromRoom, err := h.GetOneUserFromRoom(room.ID, user.ID)
			if err != nil {
				h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
					"message": "Internal Server Error DB",
				})
				return
			}
			if fromRoom.Username != "" {
				h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Hi " + username + "Welcome back in your room"})
				return
			}
			err = h.Store.AddUserToRoom(room.ID, user.ID)
			if err != nil {
				return
			}
			h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "joined this room " + room.Name})
		} else {
			h.jsonResponse(w, http.StatusUnauthorized, map[string]interface{}{"error": "Unauthorized"})
		}
	}
}

func (h *Handler) GetRooms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rooms, err := h.Store.GetRooms()
		if err != nil {
			// Handle database error
			h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"message": "Internal Server Error",
			})
			return
		}

		h.jsonResponse(w, http.StatusOK, rooms)
	}
}

func (h *Handler) CreateRoomHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		if username, ok := claims["username"].(string); ok {
			_, err := h.Store.GetUserByUsername(username)
			if err != nil {
				// Handle database error
				h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
					"message": "Internal Server Error",
				})
				return
			}
			roomName := r.FormValue("roomName")
			description := r.FormValue("description")
			roomId, err := h.Store.AddRoom(RoomItem{Name: roomName, Description: description})
			if err != nil {
				// Handle database error
				h.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
					"message": "Internal Server Error",
				})
				return
			}
			response := map[string]interface{}{"id": roomId, "name": roomName, "description": description}
			h.jsonResponse(w, http.StatusOK, response)
		} else {
			h.jsonResponse(w, http.StatusUnauthorized, map[string]interface{}{"error": "Unauthorized"})
		}
	}
}

func (h *Handler) DeleteRoomHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		QueryId := chi.URLParam(request, "id")

		id, _ := strconv.Atoi(QueryId)

		err := h.Store.DeleteRoomById(id)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		h.jsonResponse(writer, http.StatusOK, map[string]interface{}{"message": "Room " + strconv.Itoa(id) + " deleted"})
		http.Redirect(writer, request, "/chat", http.StatusSeeOther)

	}
}

func (h *Handler) UpdateRoomHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		name := r.FormValue("name")
		roomID := r.FormValue("id")
		description := r.FormValue("description")
		id, _ := strconv.Atoi(roomID)

		err := h.Store.UpdateRoom(RoomItem{ID: id, Name: name, Description: description})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "Salle modifiée", "name": name, "theme": description})

	}
}
