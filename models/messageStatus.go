package models

import (
    "time"

    "gorm.io/gorm"
)

type MessageStatus struct {
    MessageID string         `gorm:"type:char(36);primaryKey"`
    UserID    string         `gorm:"type:char(36);primaryKey"`
    IsRead    bool           `gorm:"default:false"`
    ReadAt    *time.Time     `gorm:""`
    DeletedAt gorm.DeletedAt `gorm:"index"`

    Message Message `gorm:"foreignKey:MessageID"`
    User    User    `gorm:"foreignKey:UserID"`
}
