package main

import (
	"context"
	"finance/config"
	"finance/internal/handler"
	"finance/internal/repository"
	"finance/internal/services"
	"finance/migrations"
	"finance/pkg/logger"
	"finance/pkg/postgres"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "finance/docs"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title           Finance System API
// @version         1.0
// @description     API for finance simulation system.

// @host            localhost:8181
// @BasePath        /
// @schemes   http https
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.New(cfg.LogLevel, cfg.ServiceName, cfg.AppVersion)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

	db, err := postgres.New(context.Background(), cfg.DSN, l.Logger)
	if err != nil {
		l.Logger.Fatal("failed connection to db", zap.Error(err))
	}
	defer db.Close()

	err = migrations.RunMigrations(cfg.DSN)
	if err != nil {
		l.Logger.Fatal("failed to run migrations", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db.Pool)
	limitRepo := repository.NewLimitRepository(db.Pool)
	tenorRepo := repository.NewTenorRepository(db.Pool)
	facilityRepo := repository.NewFacilityRepository(db.Pool)
	detailRepo := repository.NewDetailRepository(db.Pool)
	trx := postgres.NewTransaction(db.Pool)

	svc := services.NewService(
		userRepo,
		limitRepo,
		tenorRepo,
		facilityRepo,
		detailRepo,
		l,
		trx,
	)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notpast", validateDateNotPast)
	}

	handler := handler.NewHandler(svc, l)
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/limits", handler.ListUserLimit)
	r.GET("/tenors", handler.TenorList)
	r.POST("/calculate-installments", handler.Installment)
	r.POST("/submit-financing", handler.Submit)

	r.Run(fmt.Sprintf(":%d", cfg.HttpPort))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HttpPort),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Logger.Fatal("listen error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	l.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	l.Logger.Info("Server exiting")
}

var validateDateNotPast validator.Func = func(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	layout := "2006-01-02"

	inputDate, err := time.Parse(layout, dateStr)
	if err != nil {
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	return !inputDate.Before(today)
}
