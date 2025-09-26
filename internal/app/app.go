package app

import (
	"fmt"
	"os"
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/api/handlers"
	"github.com/Owouwun/effectivemobiletest/internal/core/api/middleware"
	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	repository_services "github.com/Owouwun/effectivemobiletest/internal/core/repository/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	dbConnectionTimeout = 30 * time.Second
)

func ConfigLogging() {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableQuote: true,
		})
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("Log level: %s", logrus.GetLevel().String())
}

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
	logrus.Info("Preparing database...")
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

	logrus.Info("Successful database preparing!")
	return db, nil
}

func PrepareRouter(db *gorm.DB) *gin.Engine {
	logrus.Info("Preparing routers...")

	router := gin.Default()

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		gin.SetMode(gin.DebugMode)
		router.Use(middleware.DebugRequestLogger())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	serviceRepo := repository_services.NewServiceRepository(db)
	subscriptionService := services.NewSubscriptionService(serviceRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService)

	apiOrders := router.Group("/service")
	{
		apiOrders.POST("", subscriptionHandler.CreateService)
		apiOrders.GET("/:id", subscriptionHandler.GetService)
		apiOrders.GET("", subscriptionHandler.GetServices)
		apiOrders.PATCH("/:id", subscriptionHandler.UpdateService)
		apiOrders.DELETE("/:id", subscriptionHandler.DeleteService)
		apiOrders.GET("/cumulate", subscriptionHandler.CumulateServices)
	}

	logrus.Info("Successful routers preparing!")
	return router
}
