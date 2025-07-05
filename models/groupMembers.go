package models

import (
    "time"

    "gorm.io/gorm"
)

type GroupMember struct {
    GroupID  string         `gorm:"type:char(36);primaryKey"`
    UserID   string         `gorm:"type:char(36);primaryKey"`
    JoinedAt time.Time      `gorm:"autoCreateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`

    User  User      `gorm:"foreignKey:UserID"`
    Group ChatGroup `gorm:"foreignKey:GroupID"`
}
