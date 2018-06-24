package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func main() {
	dbConfig := os.Getenv("TEST_CONFIG_DB")
	if dbConfig == "" {
		log.Fatal("please, set TEST_CONFIG_DB variable before running the service")
	}
	addr := os.Getenv("TEST_CONFIG_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	db, err := gorm.Open("postgres", dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	r := gin.Default()
	r.POST("/", newConfigServer(db).handle)

	log.Printf("listening at '%v'...", addr)
	r.Run(addr)
}
