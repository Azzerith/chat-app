package controllers

import (
    "errors"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "chat-app/models"
)

type MessageController struct {
    DB *gorm.DB
}

func NewMessageController(db *gorm.DB) *MessageController {
    return &MessageController{DB: db}
}

type sendMsgInput struct {
    Content    string  `json:"content" binding:"required"`
    GroupID    *string `json:"group_id"`    // optional
    ReceiverID *string `json:"receiver_id"` // optional
}


// SendMessage (POST /api/messages)
// Supports: 1) pesan ke grup (isi group_id) 2) pesan 1‑on‑1 (isi receiver_id)
func (mc *MessageController) SendMessage(c *gin.Context) {
    var input sendMsgInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if input.GroupID == nil && input.ReceiverID == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "either group_id or receiver_id required"})
        return
    }

    senderID, _ := c.Get("userID")

    msg := models.Message{
        ID:         uuid.NewString(),
        SenderID:   senderID.(string),
        GroupID:    input.GroupID,
        ReceiverID: input.ReceiverID,
        Content:    input.Content,
    }

    tx := mc.DB.Begin()
    if err := tx.Create(&msg).Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // create MessageStatus entries for recipients (unread)
    var recipients []string

    if input.GroupID != nil {
        var members []models.GroupMember
        if err := tx.Where("group_id = ?", *input.GroupID).Find(&members).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        for _, m := range members {
            if m.UserID != senderID {
                recipients = append(recipients, m.UserID)
            }
        }
    } else if input.ReceiverID != nil {
        recipients = append(recipients, *input.ReceiverID)
    }

    for _, rid := range recipients {
        tx.Create(&models.MessageStatus{
            MessageID: msg.ID,
            UserID:    rid,
            IsRead:    false,
        })
    }

    tx.Commit()

    c.JSON(http.StatusCreated, gin.H{"message_id": msg.ID})
}

// GetMessages (GET /api/messages)
// query params: group_id OR receiver_id (one‑on‑one). Order by SentAt asc.
func (mc *MessageController) GetMessages(c *gin.Context) {
    groupID := c.Query("group_id")
    recvID := c.Query("receiver_id")
    userID, _ := c.Get("userID")

    var msgs []models.Message

    q := mc.DB.Preload("Sender").Order("sent_at asc")

    switch {
    case groupID != "":
        q = q.Where("group_id = ?", groupID)
    case recvID != "":
        // ambil chat 1‑on‑1 (pesan yg saya kirim atau terima)
        q = q.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, recvID, recvID, userID)
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "group_id or receiver_id query param required"})
        return
    }

    if err := q.Find(&msgs).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // hide password inside Sender
    for i := range msgs {
        msgs[i].Sender.Password = ""
    }

    c.JSON(http.StatusOK, msgs)
}

// MarkRead (POST /api/messages/:id/read) — current user marks msg read
func (mc *MessageController) MarkRead(c *gin.Context) {
    msgID := c.Param("id")
    userID, _ := c.Get("userID")

    now := time.Now()
    res := mc.DB.Model(&models.MessageStatus{}).
        Where("message_id = ? AND user_id = ?", msgID, userID).
        Updates(map[string]interface{}{"is_read": true, "read_at": &now})

    if res.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
        return
    }
    if res.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "status not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

// DeleteMessage (DELETE /api/messages/:id) — only sender can soft‑delete
func (mc *MessageController) DeleteMessage(c *gin.Context) {
    msgID := c.Param("id")
    userID, _ := c.Get("userID")

    var msg models.Message
    if err := mc.DB.First(&msg, "id = ?", msgID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if msg.SenderID != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "only sender can delete message"})
        return
    }

    if err := mc.DB.Delete(&msg).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}
