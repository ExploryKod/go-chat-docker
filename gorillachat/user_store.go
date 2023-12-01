package main

import (
	"database/sql"
)

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db,
	}
}

type UserStore struct {
	*sql.DB
}

func (t *UserStore) GetUsers() ([]UserItem, error) {
	var users []UserItem

	rows, err := t.Query("SELECT id, username, password, admin, email FROM Users")
	if err != nil {
		return []UserItem{}, err
	}

	defer rows.Close()

	for rows.Next() {
		var user UserItem
		if err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.Admin, &user.Email); err != nil {
			return []UserItem{}, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return []UserItem{}, err
	}

	return users, nil
}

func (t *UserStore) GetUserByUsername(username string) (UserItem, error) {
	var user UserItem

	err := t.QueryRow("SELECT id, username, password, admin, email FROM Users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Admin, &user.Email)

	if err == sql.ErrNoRows {
		// User not found
		return UserItem{}, nil
	} else if err != nil {
		// Handle other database errors
		return UserItem{}, err
	}

	return user, nil
}

func (t *UserStore) GetUserById(id int) (UserItem, error) {
	var user UserItem

	err := t.QueryRow("SELECT id, username, password, admin, email FROM Users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Password, &user.Admin, &user.Email)

	if err == sql.ErrNoRows {
		// User not found
		return UserItem{}, nil
	} else if err != nil {
		// Handle other database errors
		return UserItem{}, err
	}

	return user, nil
}

func (t *UserStore) AddUser(item UserItem) (int, error) {
	res, err := t.DB.Exec("INSERT INTO Users (username, password, email, admin) VALUES (?, ?, ?, ?)", item.Username, item.Password, item.Email, item.Admin)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (t *UserStore) UpdateUser(item UserItem) error {

	_, err := t.DB.Exec("UPDATE Users SET username = ?, admin = ?, email = ? WHERE id = ?", item.Username, item.Admin, item.Email, item.ID)
	if err != nil {
		return err
	}

	return nil

}

func (t *UserStore) UpdateUserPassword(item UserItem) error {

	_, err := t.DB.Exec("UPDATE Users SET password = ? WHERE id = ?", item.Password, item.ID)
	if err != nil {
		return err
	}

	return nil

}

func (t *UserStore) DeleteUserById(id int) error {
	_, err := t.DB.Exec("DELETE FROM Users WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
