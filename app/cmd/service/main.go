package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PKSonlem/testtask-secunda-api/internal/config"
	"github.com/PKSonlem/testtask-secunda-api/internal/handler"
	"github.com/PKSonlem/testtask-secunda-api/internal/infrastructure"
	mysqlrepo "github.com/PKSonlem/testtask-secunda-api/internal/repository/mysql"
	rediscache "github.com/PKSonlem/testtask-secunda-api/internal/repository/redis"
	"github.com/PKSonlem/testtask-secunda-api/internal/service"
	"github.com/PKSonlem/testtask-secunda-api/pkg/logger"
	"github.com/PKSonlem/testtask-secunda-api/pkg/metrics"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level)
	slog.SetDefault(log) // use configured JSON handler for all slog calls

	db, err := infrastructure.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Error("connect mysql", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	rdb := infrastructure.NewRedis(cfg.Redis)
	defer rdb.Close()

	metrics.Init()

	txMgr := mysqlrepo.NewTxManager(db)

	userRepo := mysqlrepo.NewUserRepository(db)
	teamRepo := mysqlrepo.NewTeamRepository(db, txMgr)
	taskRepo := mysqlrepo.NewTaskRepository(db, txMgr)
	cache := rediscache.NewCache(rdb)
	emailClient := infrastructure.NewEmailCircuitBreaker(cfg.Email)

	authSvc := service.NewAuthService(userRepo, cfg.JWT)
	teamSvc := service.NewTeamService(teamRepo, userRepo, emailClient)
	taskSvc := service.NewTaskService(taskRepo, teamRepo, cache)

	router := handler.NewRouter(authSvc, teamSvc, taskSvc, rdb)

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("server started", "addr", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "err", err)
	}
	log.Info("bye")
}
