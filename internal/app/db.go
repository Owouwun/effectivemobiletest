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
	"github.com/sirupsen/logrus"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func waitForDBReady(dbConn string) error {
	logrus.Info("Waiting for database to be ready...")
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
				logrus.Debugf("error opening database connection: %v", err)
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
	logrus.Info("Connecting to the PostgreSQL database...")
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

	logrus.Info("Successful connecting to the database!")
	return db, nil
}

func runMigrations(dbConn string) error {
	migrationsTable := os.Getenv("MIGRATIONS_TABLE")
	if migrationsTable == "" {
		logrus.Info("Migration table was not set, use default")
		migrationsTable = "schema_migrations_effectivemobiletest"
	}

	wd, _ := os.Getwd()
	migrationsPath := "file://" + filepath.Join(wd, "migrations")
	logrus.Infof("Running database migrations; migrationsPath=%s migrationsTable=%s", migrationsPath, migrationsTable)

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
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logrus.Info("Migrations: no change (already up-to-date)")
		} else {
			return fmt.Errorf("m.Up failed: %w", err)
		}
	} else {
		logrus.Info("Migrations applied successfully!")
	}

	if v, dirty, err := m.Version(); err == nil {
		logrus.Debugf("migration version=%d dirty=%v", v, dirty)
	} else {
		rows, err := sqlDB.Query(fmt.Sprintf("SELECT version FROM %s ORDER BY version", migrationsTable))
		if err == nil {
			logrus.Debug("Applied migration versions:")
			for rows.Next() {
				var vv string
				_ = rows.Scan(&vv)
				logrus.Debug(" -", vv)
			}
			rows.Close()
		} else {
			logrus.Warnf("can't read %s: %v", migrationsTable, err)
		}
	}

	return nil
}
