package main

type UserItem struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Admin    int    `json:"admin"`
	Email    string `json:"email"`
}

type RoomItem struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Clients     map[string]*Client `json:"-"`
}

type MessageItem struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	Username  string `json:"username"`
	UserID    int    `json:"user_id"`
	RoomID    int    `json:"room_id"`
	CreatedAt string `json:"created_at"`
}

type UserStoreInterface interface {
	AddUser(item UserItem) (int, error)
	GetUserByUsername(username string) (UserItem, error)
	GetUsers() ([]UserItem, error)
	AddRoom(item RoomItem) (int, error)
	GetRoomByName(name string) (RoomItem, error)
	GetRoomById(id int) (RoomItem, error)
	DeleteRoomById(id int) error
	AddUserToRoom(roomID int, userID int) error
	GetUsersFromRoom(roomID int) ([]UserItem, error)
	GetOneUserFromRoom(roomID int, userID int) (UserItem, error)
	GetRooms() ([]RoomItem, error)
	UpdateRoom(item RoomItem) error
	DeleteUserById(id int) error
	UpdateUser(item UserItem) error
	UpdateUserPassword(item UserItem) error
	AddMessage(item MessageItem) (int, error)
	GetMessagesFromRoom(id int) ([]MessageItem, error)
	DeleteMessagesByRoomId(id int) error
	CountMessagesSent() (int, error)
}
