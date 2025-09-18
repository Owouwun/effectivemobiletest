package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/Owouwun/effectivemobiletest/cmd/docs"
	"github.com/Owouwun/effectivemobiletest/internal/app"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
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
	dbConn, err := app.BuildDBConnFromConfig()
	if err != nil {
		log.Fatalf("Failed to build DB connection string: %v", err)
	}

	db, err := app.PrepareDB(dbConn)
	if err != nil {
		log.Fatalf("Failed to prepare database: %v", err)
	}

	router := app.PrepareRouter(db)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
		log.Printf("APP_PORT not set, using default: %s", appPort)
	}

	log.Printf("Starting server on :%s", appPort)
	if err := router.Run(fmt.Sprintf(":%s", appPort)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
