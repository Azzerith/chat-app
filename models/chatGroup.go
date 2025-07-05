package models

import (
    "time"

    "gorm.io/gorm"
)

type ChatGroup struct {
    ID        string         `gorm:"type:char(36);primaryKey"`
    Name      string         `gorm:"type:varchar(100);not null"`
    CreatedBy string         `gorm:"type:char(36);not null"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`

    Members []GroupMember `gorm:"foreignKey:GroupID"`
}
