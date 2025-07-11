package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "chat-app/controllers"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
    uc := controllers.NewUserController(db)

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
    }
}