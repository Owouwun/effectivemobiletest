package helpers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Owouwun/effectivemobiletest/internal/core/api/handlers"
	"github.com/Owouwun/effectivemobiletest/internal/core/api/middleware"
	"github.com/Owouwun/effectivemobiletest/internal/core/logic/services"
	repository_services "github.com/Owouwun/effectivemobiletest/internal/core/repository/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type App struct {
	router          *gin.Engine
	db              *gorm.DB
	srv             *http.Server
	shutdownTimeout time.Duration

	sigCh chan os.Signal
}

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

func GetAppAddr() string {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
		logrus.Infof("APP_PORT is not set, using default: %s", appPort)
	}
	return ":" + appPort
}

func GetShutdownTimeout() time.Duration {
	if s := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); s != "" {
		if secs, err := strconv.Atoi(s); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
		logrus.Warnf("Invalid SHUTDOWN_TIMEOUT_SECONDS=%s, using default", s)
	}
	return 30 * time.Second
}

func NewApp(router *gin.Engine, db *gorm.DB, addr string, shutdownTimeout time.Duration) *App {
	return &App{
		router:          router,
		db:              db,
		srv:             newHTTPServer(addr, router),
		shutdownTimeout: shutdownTimeout,
		sigCh:           make(chan os.Signal, 1),
	}
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (a *App) Run(ctx context.Context) error {
	listenErrCh := a.startServer()

	err := a.waitForSignalOrListenError(listenErrCh)
	if err != nil {
		return fmt.Errorf("startup error: %w", err)
	}

	if err := a.shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	return nil
}

func (a *App) waitForSignalOrListenError(listenErrCh chan error) error {
	signal.Notify(a.sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-a.sigCh:
		logrus.Infof("Received signal %s, beginning shutdown", sig.String())
		return nil
	case err := <-listenErrCh:
		if err != nil {
			logrus.Errorf("Listen error: %v", err)
			return err
		}
		logrus.Info("Listen finished without error")
		return nil
	}
}

func (a *App) shutdown(parentCtx context.Context) error {
	ctx, cancel := context.WithTimeout(parentCtx, a.shutdownTimeout)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		logrus.Errorf("HTTP server graceful shutdown error: %v", err)
		inErr := a.srv.Close()
		if inErr != nil {
			logrus.Errorf("HTTP server hard shutdown error: %v", inErr)
		}
	}

	if err := closeDB(a.db); err != nil {
		logrus.Errorf("DB close error: %v", err)
		return err
	}

	logrus.Info("Application shutdown complete")
	return nil
}

func (a *App) startServer() chan error {
	listenErrCh := make(chan error, 1)
	go func() {
		logrus.Infof("Starting HTTP server on %s", a.srv.Addr)
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErrCh <- err
		}
		close(listenErrCh)
	}()
	return listenErrCh
}
