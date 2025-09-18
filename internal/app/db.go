package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func waitForDBReady(dbConn string) error {
	log.Println("Waiting for database to be ready...")
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectionTimeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			db, err := sql.Open("postgres", dbConn)
			if err != nil {
				log.Printf("error opening database connection: %v", err)
				continue
			}
			if err := db.Ping(); err == nil {
				db.Close()
				return nil
			}
			db.Close()
		}
	}
}

func connectToDB(dbConn string) (*gorm.DB, error) {
	log.Println("Connecting to the PostgreSQL database...")
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	db, err := gorm.Open(gorm_postgres.Open(dbConn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successful connection to the database!")
	return db, nil
}

func runMigrations(dbConn string) error {
	migrationsTable := os.Getenv("MIGRATIONS_TABLE")
	if migrationsTable == "" {
		migrationsTable = "schema_migrations_effectivemobiletest"
	}

	wd, _ := os.Getwd()
	migrationsPath := "file://" + filepath.Join(wd, "migrations")
	log.Printf("Running database migrations; migrationsPath=%s migrationsTable=%s\n", migrationsPath, migrationsTable)

	sqlDB, err := sql.Open("postgres", dbConn)
	if err != nil {
		return fmt.Errorf("sql.Open failed: %w", err)
	}
	defer sqlDB.Close()

	cfg := &postgres.Config{
		MigrationsTable: migrationsTable,
	}

	driver, err := postgres.WithInstance(sqlDB, cfg)
	if err != nil {
		return fmt.Errorf("postgres.WithInstance failed: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance failed: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Migrations: no change (already up-to-date)")
		} else {
			return fmt.Errorf("m.Up failed: %w", err)
		}
	} else {
		log.Println("Migrations applied successfully!")
	}

	if v, dirty, err := m.Version(); err == nil {
		log.Printf("migration version=%d dirty=%v\n", v, dirty)
	} else {
		rows, err := sqlDB.Query(fmt.Sprintf("SELECT version FROM %s ORDER BY version", migrationsTable))
		if err == nil {
			log.Println("Applied migration versions:")
			for rows.Next() {
				var vv string
				_ = rows.Scan(&vv)
				log.Println(" -", vv)
			}
			rows.Close()
		} else {
			log.Printf("can't read %s: %v", migrationsTable, err)
		}
	}

	return nil
}
