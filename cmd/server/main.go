package main

import (
	"context"

	_ "github.com/Owouwun/effectivemobiletest/cmd/docs"
	"github.com/Owouwun/effectivemobiletest/internal/app"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// @title           Effective Mobile Test Assignment
// @version         1.0
// @description     This is a test assignment for Effective Mobile team.
// @contact.name    Ivan Kuznetsov
// @contact.email   kuznetsovivangio@gmail.com
// @host            localhost:8080
// @BasePath        /service

func main() {
	if err := app.Run(context.Background()); err != nil {
		logrus.Fatalf("Application stopped with error: %v", err)
	}
}
