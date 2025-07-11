package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string         `gorm:"type:char(36);primaryKey"`
	Username  string         `gorm:"type:varchar(50);not null;unique"`
	Email     string         `gorm:"type:varchar(100);not null;unique"`
	Password  string         `gorm:"type:varchar(255);not null"`
	IsOnline  bool           `gorm:"type:boolean;not null;default:false"`
	LastSeen  *time.Time     `gorm:""`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // soft delete
}
