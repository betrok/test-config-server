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

	switch len(os.Args) {
	case 1:
		run(db, addr)

	case 2:
		switch os.Args[1] {
		case "run":
			run(db, addr)

		case "migrate":
			migrate(db)

		case "rollback":
			log.Println("destinnation_migration_id required for rollback")
			help()

		default:
			help()
		}

	case 3:
		if os.Args[1] != "rollback" {
			help()
		}
		rollback(db, os.Args[2])

	default:
		help()
	}
}

func help() {
	log.Printf(`Usage: %v [command] [destinnation_migration_id]
	Commands:
		run (default)  just start the service
		migrate        perform all missing migrations
		rollback       rollback up to destinnation_migration_id`, os.Args[0])
	os.Exit(1)
}

func run(db *gorm.DB, addr string) {
	err := ensureMigration(db)
	if err != nil {
		log.Printf("migrations in the database do not match expections: %v", err)
		log.Printf("did you run `%v migrate` before running the service?", os.Args[0])
		os.Exit(1)
	}

	r := gin.Default()
	r.POST("/", newConfigServer(db).handle)
	err = r.Run(addr)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
