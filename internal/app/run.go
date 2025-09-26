package app

import (
	"context"
	"fmt"

	helpers "github.com/Owouwun/effectivemobiletest/internal/app/helpers"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run(ctx context.Context) error {
	helpers.ConfigLogging()

	dbConn, err := helpers.BuildDBConnFromConfig()
	if err != nil {
		return fmt.Errorf("failed to build DB connection string: %w", err)
	}

	db, err := helpers.PrepareDB(dbConn)
	if err != nil {
		return fmt.Errorf("failed to prepare database: %w", err)
	}

	router := helpers.PrepareRouter(db)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	addr := helpers.GetAppAddr()
	shutdownTimeout := helpers.GetShutdownTimeout()
	a := helpers.NewApp(router, db, addr, shutdownTimeout)

	logrus.Infof("App constructed; delegating run to App.Run")
	return a.Run(ctx)
}
