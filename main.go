package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"chat-app/models"
	"chat-app/routes"
)

func main() {
	// 1. Koneksi DB
	dsn := getDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal konek ke database: %v", err)
	}

	// 2. Auto‑migrate
	if err := db.AutoMigrate(
		&models.User{},
		&models.ChatGroup{},
		&models.GroupMember{},
		&models.Message{},
		&models.MessageStatus{},
	); err != nil {
		log.Fatalf("Gagal migrate tabel: %v", err)
	}
	fmt.Println("Database terkoneksi & migrasi selesai.")

	// 3. Setup Gin & routes
	router := gin.Default()
	routes.RegisterRoutes(router, db)
	router.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 4. Jalankan server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Gagal start server: %v", err)
	}
}

func getDSN() string {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	dbname := os.Getenv("DB_NAME")

	if user == "" {
		user = "root"
	}
	if pass == "" {
		pass = ""
	}
	if host == "" {
		host = "127.0.0.1:3306"
	}
	if dbname == "" {
		dbname = "chatdb"
	}

	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, dbname)
}
