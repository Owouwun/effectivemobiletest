package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/api/handlers"
	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	repository_services "github.com/Owouwun/effectivemobiletest/internal/core/repository/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	dbConnectionTimeout = 30 * time.Second
)

func BuildDBConnFromConfig() (string, error) {
	conn := os.Getenv("DATABASE_CONN")
	if conn != "" {
		return conn, nil
	}
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("POSTGRES_DB")
	if user == "" || pass == "" || host == "" || dbName == "" {
		return "", fmt.Errorf("missing DB secret envs")
	}

	conn = fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, host, dbName)
	return conn, nil
}

func PrepareDB(dbConn string) (*gorm.DB, error) {
	log.Println("Preparing database...")
	if err := waitForDBReady(dbConn); err != nil {
		return nil, fmt.Errorf("failed to wait for database: %w", err)
	}

	if err := runMigrations(dbConn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	db, err := connectToDB(dbConn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func PrepareRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	serviceRepo := repository_services.NewServiceRepository(db)
	subscriptionService := services.NewSubscriptionService(serviceRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)

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
