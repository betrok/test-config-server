package migration

import (
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

// @TODO Migrations table prefix?
// @TODO Forward migration to passed destination level.

// Migration represents a description and logic of single migration step.
type Migration struct {
	// Non-empty unique key
	ID string `gorm:"primary_key"`
	// Human-redable description
	Description string
	// Current time will be set here before saving migration info to db after execution.
	PerformedAt time.Time

	Rerform  func(tx *gorm.DB) error `gorm:"-"`
	Rollback func(tx *gorm.DB) error `gorm:"-"`
}

// baseMigration will be added to the top if any migration list to prepare the shema of internal data.
var baseMigration = Migration{
	ID:          "0000_migrations_table",
	Description: "creates table with migration data",
	Rerform: func(tx *gorm.DB) error {
		return tx.CreateTable(&Migration{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable(&Migration{}).Error
	},
}

// Migrate applies all missing migrations to db in the order determined by the argument slice.
// Function aborts if any unknown migrations are presented in db.
func Migrate(db *gorm.DB, migrations []Migration) error {
	migrations = append([]Migration{baseMigration}, migrations...)
	tx := db.Begin()

	performed, err := loadPerformedMigrations(tx, migrations)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, mig := range migrations {
		// already performed
		if performed[mig.ID] {
			continue
		}
		log.Printf("applying migration '%v'(%v)...", mig.ID, mig.Description)

		err := mig.Rerform(tx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to perform migration '%v': %v", mig.ID, err)
		}

		mig.PerformedAt = time.Now()
		err = tx.Save(&mig).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save migration info: %v", err)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit after migration complete: %v", err)
	}

	return nil
}

// Rollback rolls back the migrations up to destMigrationLvl exclusive.
// With "" as dest destMigrationLvl will roll back all migrations.
// Function aborts if any unknown migrations are presented in db.
func Rollback(db *gorm.DB, migrations []Migration, destMigrationLvl string) error {
	migrations = append([]Migration{baseMigration}, migrations...)

	dest := -1

	if destMigrationLvl != "" {
		for i, mig := range migrations {
			if mig.ID == destMigrationLvl {
				dest = i
				break
			}
		}
		if dest == -1 {
			return fmt.Errorf("unknown destination migration level '%v'", destMigrationLvl)
		}
	}

	tx := db.Begin()

	performed, err := loadPerformedMigrations(tx, migrations)
	if err != nil {
		tx.Rollback()
		return err
	}

	for i := len(migrations) - 1; i > dest; i-- {
		mig := migrations[i]
		if !performed[mig.ID] {
			continue
		}

		log.Printf("rolling back migration '%v'(%v)...", mig.ID, mig.Description)

		err := mig.Rollback(tx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to rollback migration '%v': %v", mig.ID, err)
		}

		// It was BaseMigration. There is no more migrations table db...
		if i == 0 {
			break
		}
		err = tx.Delete(&mig).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove rollbacked migration info: %v", err)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("failed to commit after rollback complete: %v", err)
	}

	return nil
}

// Ensure whether the database corresponds to migrations.
func Ensure(db *gorm.DB, migrations []Migration) error {
	migrations = append([]Migration{baseMigration}, migrations...)

	performed, err := loadPerformedMigrations(db, migrations)
	if err != nil {
		return err
	}

	for _, mig := range migrations {
		if !performed[mig.ID] {
			return fmt.Errorf("db does not contain all expected migrations(first missing is '%v')", mig.ID)
		}
	}

	return nil
}

// loadPerformedMigrations returns 'set' of previously performed migrations
// and returns error in case of any unknown migrations(not listed in the argument).
// Passed migration list shuold contain the base migration as well.
func loadPerformedMigrations(tx *gorm.DB, migrations []Migration) (map[string]bool, error) {
	if !tx.HasTable(&Migration{}) {
		return map[string]bool{}, nil
	}

	var performed []Migration
	err := tx.Find(&performed).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load perofrmed migrations: %v", err)
	}

	ret := make(map[string]bool)
	for _, old := range performed {
		known := false
		for _, mig := range migrations {
			if mig.ID == old.ID {
				known = true
				ret[old.ID] = true
				break
			}
		}
		if !known {
			return nil, fmt.Errorf("unknown migration '%v' was found in db", old.ID)
		}
	}

	return ret, nil
}
