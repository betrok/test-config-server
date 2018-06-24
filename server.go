package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type configServer struct {
	db *gorm.DB
}

// Config represents the associated structure in the database.
// It has composed (type, name) primary key, the configuration data itself is stored in the jsonb form.
type Config struct {
	Type string `gorm:"primary_key"`
	Name string `gorm:"primary_key"`
	Data postgres.Jsonb
}

func newConfigServer(db *gorm.DB) *configServer {
	return &configServer{db: db}
}

func (s configServer) handle(c *gin.Context) {
	var request struct {
		Type string
		Name string `json:"Data"`
	}

	err := c.BindJSON(&request)
	if err != nil {
		log.Printf("failed to decode request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bad request",
		})
		return
	}

	if request.Type == "" || request.Name == "" {
		log.Println("incomplite request: empty type or data")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "empty type or data",
		})
		return
	}

	config := Config{
		Type: request.Type,
		Name: request.Name,
	}

	res := s.db.First(&config)
	switch {
	case res.RecordNotFound():
		log.Printf("config '%v' with type '%v' not found", request.Name, request.Type)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "record not found",
		})
		return
	case res.Error != nil:
		log.Printf("failed to load config data: %v", res.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "db error",
		})
		return
	default:
		c.JSON(http.StatusOK, config.Data.RawMessage)
		return
	}
}
