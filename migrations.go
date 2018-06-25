package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/betrok/test-config-server/migration"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

var migrations = []migration.Migration{
	{
		ID:          "0010_configs_table",
		Description: "creates table with config data",
		Rerform: func(tx *gorm.DB) error {
			return tx.Exec(`
				CREATE TABLE "configs" (
					"type" text,
					"name" text,
					"data" jsonb,
					PRIMARY KEY ("type","name")
				)`).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable(&Config{}).Error
		},
	},
	// Should this really be in migrations?..
	// It was intended to use an _external_ migration module, i guess.
	{
		ID:          "0020_test_config_data",
		Description: "fills db with the test data",
		Rerform: func(tx *gorm.DB) error {
			for _, conf := range testData {
				err := tx.Create(&conf).Error
				if err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			for _, conf := range testData {
				err := tx.Delete(&conf).Error
				if err != nil {
					return err
				}
			}
			return nil
		},
	},
}

func toJsonb(str string) postgres.Jsonb {
	return postgres.Jsonb{
		RawMessage: json.RawMessage(str),
	}
}

var testData = []Config{
	{
		Type: "database.postgres",
		Name: "service.test",
		Data: toJsonb(`
		{
			"host": "localhost",
			"port": "5432",
			"database": "devdb",
			"user": "mr_robot",
			"password": "secret",
			"schema": "public"
		}`),
	},
	{
		Type: "rabbit.log",
		Name: "service.test",
		Data: toJsonb(`
		{
			"host": "10.0.5.42",
			"port": "5671",
			"virtualhost": "/",
			"user": "guest",
			"password": "guest"
		}`),
	},
}

func migrate(db *gorm.DB) {
	err := migration.Migrate(db, migrations)
	if err != nil {
		log.Printf("migration failed: %v", err)
		os.Exit(1)
	} else {
		log.Println("migration finished")
		os.Exit(0)
	}
}

func rollback(db *gorm.DB, dest string) {
	err := migration.Rollback(db, migrations, dest)
	if err != nil {
		log.Printf("rollback failed: %v", err)
		os.Exit(1)
	} else {
		log.Println("rollback finished")
		os.Exit(0)
	}
}

func ensureMigration(db *gorm.DB) error {
	return migration.Ensure(db, migrations)
}
