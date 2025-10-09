package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

type Task struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`   // "pending", "in_progress", "completed"
	Priority    string    `json:"priority"` // "low", "medium", "high"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DashboardStats struct {
	TotalTasks        int
	PendingTasks      int
	InProgressTasks   int
	CompletedTasks    int
	HighPriorityTasks int
}
