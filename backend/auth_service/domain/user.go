package domain

import (
	"context"
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
}

type AuthUsecase interface {
	Login(ctx context.Context, email string, password string) (string, string, *User, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
}
