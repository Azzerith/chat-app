package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "chat-app/controllers"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
    uc := controllers.NewUserController(db)
	gc := controllers.NewGroupController(db)
	mc := controllers.NewMessageController(db)


    // public endpoints
    r.POST("/api/register", uc.Register)
    r.POST("/api/login", uc.Login)

    // protected endpoints
    api := r.Group("/api").Use(uc.JWTAuthMiddleware())
    {
        api.GET("/users", uc.GetUsers)
        api.GET("/users/:id", uc.GetUser)
        api.PUT("/users/:id", uc.UpdateUser)
        // api.DELETE("/users/:id", uc.DeleteUser)
        api.POST("/logout", uc.Logout)

		api.POST("/groups", gc.CreateGroup)
        api.GET("/groups", gc.GetGroups)
        api.POST("/groups/:id/join", gc.JoinGroup)
        api.POST("/groups/:id/leave", gc.LeaveGroup)
        api.DELETE("/groups/:id", gc.DeleteGroup)// delete group (creator only)

		api.POST("/messages", mc.SendMessage)
		api.GET("/messages", mc.GetMessages)
		api.POST("/messages/:id/read", mc.MarkRead)
		api.DELETE("/messages/:id", mc.DeleteMessage)
    }
}