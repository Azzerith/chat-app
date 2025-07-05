package models

import (
    "time"

    "gorm.io/gorm"
)

type Message struct {
    ID         string         `gorm:"type:char(36);primaryKey"`
    SenderID   string         `gorm:"type:char(36);not null"`
    GroupID    *string        `gorm:"type:char(36);"`       // nullable: pesan ke grup
    ReceiverID *string        `gorm:"type:char(36);"`       // nullable: pesan ke user (1-on-1)
    Content    string         `gorm:"type:text;not null"`
    SentAt     time.Time      `gorm:"autoCreateTime"`
    DeletedAt  gorm.DeletedAt `gorm:"index"`

    Sender   User       `gorm:"foreignKey:SenderID"`
    Group    ChatGroup  `gorm:"foreignKey:GroupID"`
    Receiver User       `gorm:"foreignKey:ReceiverID"`
}
