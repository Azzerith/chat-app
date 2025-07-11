package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"chat-app/models"
)

type GroupController struct {
	DB *gorm.DB
}

func NewGroupController(db *gorm.DB) *GroupController {
	return &GroupController{DB: db}
}

// CreateGroup (POST /api/groups)
func (gc *GroupController) CreateGroup(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	group := models.ChatGroup{
		ID:        uuid.NewString(),
		Name:      input.Name,
		CreatedBy: userID.(string),
	}

	if err := gc.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	member := models.GroupMember{
		GroupID:  group.ID,
		UserID:   userID.(string),
		JoinedAt: time.Now(),
	}
	gc.DB.Create(&member)

	c.JSON(http.StatusCreated, group)
}

// GetGroups (GET /api/groups)
func (gc *GroupController) GetGroups(c *gin.Context) {
	var groups []models.ChatGroup
	if err := gc.DB.Preload("Members").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// JoinGroup (POST /api/groups/:id/join)
func (gc *GroupController) JoinGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID, _ := c.Get("userID")

	var existing models.GroupMember
	if err := gc.DB.First(&existing, "group_id = ? AND user_id = ?", groupID, userID).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already joined"})
		return
	}

	member := models.GroupMember{
		GroupID:  groupID,
		UserID:   userID.(string),
		JoinedAt: time.Now(),
	}
	if err := gc.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "joined group"})
}

// LeaveGroup (POST /api/groups/:id/leave)
func (gc *GroupController) LeaveGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID, _ := c.Get("userID")

	if err := gc.DB.Delete(&models.GroupMember{}, "group_id = ? AND user_id = ?", groupID, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "left group"})
}

// DeleteGroup (DELETE /api/groups/:id)
func (gc *GroupController) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID, _ := c.Get("userID")

	var group models.ChatGroup
	if err := gc.DB.First(&group, "id = ?", groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	if group.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only creator can delete group"})
		return
	}

	gc.DB.Delete(&models.GroupMember{}, "group_id = ?", groupID)
	gc.DB.Delete(&group)

	c.JSON(http.StatusOK, gin.H{"message": "group deleted"})
}
