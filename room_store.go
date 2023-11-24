package main

import (
	"database/sql"
)

type RoomItem struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Clients     map[string]*Client `json:"-"`
}

func (t *UserStore) AddRoom(item RoomItem) (int, error) {
	res, err := t.DB.Exec("INSERT INTO Rooms (name, description) VALUES (?, ?)", item.Name, item.Description)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (t *UserStore) GetRoomByName(name string) (RoomItem, error) {
	var room RoomItem

	err := t.QueryRow("SELECT id, name, description FROM Rooms WHERE name = ?", name).
		Scan(&room.ID, &room.Name, &room.Description)

	if err == sql.ErrNoRows {
		// Room not found
		return RoomItem{}, nil
	} else if err != nil {
		// Handle other database errors
		return RoomItem{}, err
	}

	return room, nil
}

func (t *UserStore) GetRoomById(id int) (RoomItem, error) {
	var room RoomItem

	err := t.QueryRow("SELECT id, name, description FROM Rooms WHERE id = ?", id).
		Scan(&room.ID, &room.Name, &room.Description)

	if err == sql.ErrNoRows {
		// Room not found
		return RoomItem{}, nil
	} else if err != nil {
		// Handle other database errors
		return RoomItem{}, err
	}

	return room, nil
}

func (t *UserStore) AddUserToRoom(roomID int, userID int) error {
	_, err := t.DB.Exec("INSERT INTO User_Room (user_id, room_id) VALUES (?, ?)", userID, roomID)
	if err != nil {
		return err
	}
	return nil
}

func (t *UserStore) GetUsersFromRoom(roomID int) ([]UserItem, error) {
	var users []UserItem

	rows, err := t.Query("SELECT id, username, password FROM Users INNER JOIN User_Room ON Users.id = User_Room.user_id WHERE User_Room.room_id = ?", roomID)
	if err != nil {
		return []UserItem{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var user UserItem
		if err = rows.Scan(&user.ID, &user.Username, &user.Password); err != nil {
			return []UserItem{}, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return []UserItem{}, err
	}

	return users, nil
}

func (t *UserStore) GetOneUserFromRoom(roomID int, userID int) (UserItem, error) {
	var user UserItem

	err := t.QueryRow("SELECT Users.id, Users.username, Users.password FROM Users INNER JOIN User_Room ON Users.id = User_Room.user_id WHERE User_Room.room_id = ? AND Users.id = ? LIMIT 1", roomID, userID).
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

func (t *UserStore) GetRooms() ([]RoomItem, error) {
	var rooms []RoomItem

	rows, err := t.Query("SELECT id, name, description FROM Rooms")
	if err != nil {
		return []RoomItem{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var room RoomItem
		if err = rows.Scan(&room.ID, &room.Name, &room.Description); err != nil {
			return []RoomItem{}, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return []RoomItem{}, err
	}

	return rooms, nil
}

func (t *UserStore) UpdateRoom(item RoomItem) error {

	_, err := t.DB.Exec("UPDATE Rooms SET name = ?, description = ? WHERE id = ?", item.Name, item.Description, item.ID)
	if err != nil {
		return err
	}

	return nil

}

func (t *UserStore) DeleteRoomById(id int) error {
	_, err := t.DB.Exec("DELETE FROM Rooms WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
