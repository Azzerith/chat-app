package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"chat-app/models"
)

func main() {
	// Konfigurasi koneksi database
	dsn := getDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal konek ke database: %v", err)
	}

	// Auto migrate tabel
	err = db.AutoMigrate(
		&models.User{},
		&models.ChatGroup{},
		&models.GroupMember{},
		&models.Message{},
		&models.MessageStatus{},
	)
	if err != nil {
		log.Fatalf("Gagal migrate tabel: %v", err)
	}

	fmt.Println("Database berhasil terkoneksi dan migrasi selesai.")

	// Setup Gin
	router := gin.Default()

	router.Run()
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
