package main

import (
	"context"
	"finance/config"
	"finance/internal/handler"
	"finance/internal/repository"
	"finance/internal/services"
	"finance/pkg/logger"
	"finance/pkg/postgres"
	"fmt"
	"log"
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

// @host            localhost:8080
// @BasePath        /
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
	v1 := r.Group("/api/v1")
	{
		v1.GET("/user-limits", handler.ListUserLimit)
		v1.GET("/tenor", handler.TenorList)
		v1.POST("/calculate-installments", handler.Installment)
		v1.POST("/submit-financing", handler.Submit)
	}

	r.Run(fmt.Sprintf(":%d", cfg.HttpPort))
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
