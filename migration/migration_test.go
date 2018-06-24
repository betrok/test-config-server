package migration

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestMigration(t *testing.T) {
	type TestData struct {
		Data string `gorm:"primary_key"`
	}

	var migrations = []Migration{
		{
			ID: "schema",
			Rerform: func(tx *gorm.DB) error {
				return tx.CreateTable(&TestData{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable(&TestData{}).Error
			},
		},
		{
			ID: "fill_one",
			Rerform: func(tx *gorm.DB) error {
				return tx.Create(&TestData{
					Data: "one",
				}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Delete(&TestData{}, "data = ?", "one").Error
			},
		},
	}

	db, err := gorm.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	err = Migrate(db, migrations)
	if err != nil {
		t.Fatalf("first Migrate() failed: %v", err)
	}

	// Add extra migration and run Migrate() again.
	// If the previosly applied migrations are performed, it will fail the due unique constrait.
	migrations = append(migrations, Migration{
		ID: "fill_two",
		Rerform: func(tx *gorm.DB) error {
			return tx.Create(&TestData{
				Data: "two",
			}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Delete(&TestData{}, "data = ?", "two").Error
		},
	})
	err = Migrate(db, migrations)
	if err != nil {
		t.Fatalf("second Migrate() failed: %v", err)
	}

	if Migrate(db, []Migration{}) == nil {
		t.Fatalf("Migrate() did not fail on unknown loaded migration")
	}

	err = Ensure(db, migrations)
	if err != nil {
		t.Fatalf("Ensure() failed: %v", err)
	}

	if Ensure(db, append(migrations, Migration{ID: "something"})) == nil {
		t.Fatalf("Ensure() did not fail where it should have")
	}

	var data TestData
	err = db.Find(&data, "data = ?", "two").Error
	if err != nil {
		t.Fatalf("failed to load data after migration: %v", err)
	}

	if Rollback(db, migrations, "unknown") == nil {
		t.Fatalf("Rollback() did not fail on the unknown destination migration level")
	}

	if Rollback(db, nil, "fill_one") == nil {
		t.Fatalf("Rollback() did not fail on unknown loaded migration")
	}

	err = Rollback(db, migrations, "fill_one")
	if err != nil {
		t.Fatalf("valid Rollback() failed: %v", err)
	}

	data = TestData{}
	if !db.Find(&data, "data = ?", "two").RecordNotFound() {
		t.Fatal("canceled data still exists after Rollback()")
	}

	data = TestData{}
	err = db.Find(&data, "data = ?", "one").Error
	if err != nil {
		t.Fatalf("failed to load data after after Rollback(): %v", err)
	}

	err = Rollback(db, migrations, "")
	if err != nil {
		t.Fatalf("cleanup Rollback() failed: %v", err)
	}
}
