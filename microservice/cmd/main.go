package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"review-manager/internal/httpapi"
	"review-manager/internal/repository"
	"review-manager/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("DSN не задано в окружении")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT нужно задать в окружении")
	}

	admToken := os.Getenv("ADMIN_TOKEN")
	if admToken == "" {
		log.Fatalf("ADMIN_TOKEN надо задать в окружении")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Не получилось создать pgxpool: %v", err)
	}
	defer pool.Close()

	// Репозитории
	teamRepo := repository.NewPgTeamRepo(pool)
	userRepo := repository.NewPgUserRepo(pool)
	prRepo := repository.NewPgPrRepo(pool)

	// Менеджер транзакций
	txMgr := repository.NewPgTxManager(pool)

	// Cервисы
	teamSvc := service.NewTeamService(teamRepo, userRepo, txMgr)
	userSvc := service.NewUserService(userRepo)
	prSvc := service.NewPRService(prRepo, userRepo, txMgr)

	// HTTP API
	h := httpapi.NewHandler(teamSvc, userSvc, prSvc, admToken)
	handler := httpapi.NewMux(h)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP сервер работает по адресу %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Выключаемся...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при отключении сервера: %v", err)
	}
}
