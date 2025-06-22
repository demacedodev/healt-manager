package main

import (
	"callers-go/application"
	"callers-go/infrastructure/client"
	presentation "callers-go/infrastructure/http"
	"callers-go/infrastructure/repository"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	db, err := repository.NewPersistentStorage(&repository.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})
	if err != nil {
		panic(err)
	}

	cfg := &application.Config{
		Client: client.NewClient(&client.Config{
			CallersBaseURL: "http://127.0.0.1:8081",
			DevicesBaseURL: "http://127.0.0.1:8888",
			Timeout:        1200 * time.Second,
		}),
		Storage: db,
		Cache:   repository.NewMemoryStorage(),
	}

	task := application.NewTask(cfg)
	app := application.NewApp(cfg)
	callers := presentation.NewHandlers(app)

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health/callers", callers.GetCallers)

	cronLoad := time.NewTicker(30 * time.Second)
	defer cronLoad.Stop()
	go Runner(cronLoad, "LoadDevices", func() error {
		return task.LoadDevices()
	})

	cronUpdate := time.NewTicker(1 * time.Second)
	defer cronUpdate.Stop()
	go Runner(cronUpdate, "UpdateDevices", func() error {
		return task.UpdateStatus()
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		fmt.Println("🚀 Server running on http://127.0.0.1:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("❌ Listen error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("⚠️  Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("❌ Server forced to shutdown: %v\n", err)
	}

	fmt.Println("✅ Server exited gracefully")
}

func Runner(ticker *time.Ticker, taskName string, f func() error) {
	for {
		select {
		case <-ticker.C:
			_ = f()
		}
	}
}
