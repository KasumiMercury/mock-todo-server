package domain

import "time"

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	UserID    int    `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
