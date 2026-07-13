package db

import (
	"time"
)

type Session struct {
	SessionID string    `gorm:"primaryKey"`
	UserID      string    `gorm:"index"`
	AccessToken string
	ExpiresAt   time.Time `gorm:"index"`
	CreatedAt   time.Time
}
