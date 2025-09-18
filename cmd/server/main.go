package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/Owouwun/effectivemobiletest/cmd/docs"

	"github.com/Owouwun/effectivemobiletest/internal/core/api/handlers"
	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	repository_services "github.com/Owouwun/effectivemobiletest/internal/core/repository/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres migration
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Migrations from file
	_ "github.com/lib/pq"                                      // Register Postgres driver
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	dbConnectionTimeout = 30 * time.Second
)

// @title           Effective Mobile Test Assignment
// @version         1.0
// @description     This is a test assignment for Effective Mobile team.

// @contact.name   Ivan Kuznetsov
// @contact.email  kuznetsovivangio@gmail.com

// @host      localhost:8080
// @BasePath  /service

func main() {
	db := prepareDB()
	router := prepareRouter(db)

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func prepareRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	serviceRepo := repository_services.NewServiceRepository(db)
	subscriptionService := services.NewSubscriptionService(serviceRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiOrders := router.Group("/service")
	{
		apiOrders.POST("", subscriptionHandler.CreateService)
		apiOrders.GET("/:service_name", subscriptionHandler.GetService)
		apiOrders.GET("", subscriptionHandler.GetServices)
		apiOrders.PATCH("/:service_name", subscriptionHandler.UpdateService)
		apiOrders.DELETE("/:service_name", subscriptionHandler.DeleteService)
		apiOrders.GET("/cumulate", subscriptionHandler.CumulateServices)
	}

	return router
}

func prepareDB() *gorm.DB {
	dbConn := os.Getenv("DATABASE_CONN")
	if dbConn == "" {
		log.Fatal("DATABASE_CONN environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectionTimeout)
	defer cancel()

	if err := waitForDBReady(ctx, dbConn); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	runMigrations(dbConn)

	log.Println("Connecting to the PostgreSQL database...")
	db, err := gorm.Open(postgres.Open(dbConn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	return db
}

func runMigrations(dbConn string) {
	log.Println("Running database migrations...")

	m, err := migrate.New(
		"file://migrations", // Путь к папке с миграциями
		dbConn,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations applied successfully!")
}

func waitForDBReady(ctx context.Context, dbConn string) error {
	log.Println("Waiting for database to be ready...")

	done := make(chan error)

	go func() {
		for {
			db, err := sql.Open("postgres", dbConn)
			if err != nil {
				done <- err
				return
			}
			defer func() {
				err := db.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()

			if err := db.Ping(); err == nil {
				done <- nil
				return
			}

			// Wait till the next try
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done(): // Timeout
		return ctx.Err()
	}
}
