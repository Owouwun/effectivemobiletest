package app

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"gorm.io/driver/postgres"
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
	db, err := gorm.Open(postgres.Open(dbConn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func runMigrations(dbConn string) error {
	log.Println("Running database migrations...")
	m, err := migrate.New("file://migrations", dbConn)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		if err == migrate.ErrNoChange {
			log.Println("Nothing to migrate for database")
			return nil
		}
		return err
	}
	log.Println("Database migrations applied successfully!")
	return nil
}
