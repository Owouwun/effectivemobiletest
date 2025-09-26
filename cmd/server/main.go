package main

import (
	"fmt"
	"os"

	_ "github.com/Owouwun/effectivemobiletest/cmd/docs"
	"github.com/Owouwun/effectivemobiletest/internal/app"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Effective Mobile Test Assignment
// @version         1.0
// @description     This is a test assignment for Effective Mobile team.
// @contact.name    Ivan Kuznetsov
// @contact.email   kuznetsovivangio@gmail.com
// @host            localhost:8080
// @BasePath        /service

func main() {
	app.ConfigLogging()

	dbConn, err := app.BuildDBConnFromConfig()
	if err != nil {
		logrus.Fatalf("Failed to build DB connection string: %v", err)
	}

	db, err := app.PrepareDB(dbConn)
	if err != nil {
		logrus.Fatalf("Failed to prepare database: %v", err)
	}

	router := app.PrepareRouter(db)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
		logrus.Infof("APP_PORT is not set, using default: %s", appPort)
	}

	logrus.Infof("Starting server on :%s", appPort)
	if err := router.Run(fmt.Sprintf(":%s", appPort)); err != nil {
		logrus.Fatalf("Server failed to start: %v", err)
	}
}
