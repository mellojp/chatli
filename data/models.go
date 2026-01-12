package data

import (
	"time"
)

type Message struct {
	Id             string    `json:"id,omitempty"`
	Type           string    `json:"type"`
	UserId         string    `json:"user_id,omitempty"`
	SenderUsername string    `json:"sender_username"`
	Content        string    `json:"content"`
	SentAt         time.Time `json:"created_at,omitempty"`
	RoomId         string    `json:"room_id"`
}

type Room struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	CreatorId string     `json:"creator_id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Session struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	UserId      string `json:"user_id"`
	JoinedRooms []Room `json:"joined_rooms"`
}