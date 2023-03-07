package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/aria3ppp/url-shortener/helper"
	"github.com/aria3ppp/url-shortener/internal/core/usecase"
	"github.com/aria3ppp/url-shortener/internal/generator"
	"github.com/aria3ppp/url-shortener/internal/handler"
	"github.com/aria3ppp/url-shortener/internal/repository"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	repo := repository.NewRepository(db)
	generator := generator.NewRandomStringGenerator(6)

	serviceUseCases := usecase.NewService(repo, generator)

	handler := handler.NewHandler(serviceUseCases)

	router := echo.New()
	helper.HandleRoutes(router, handler)

	if err := router.Start(":" + os.Getenv("SERVER_PORT")); err != nil {
		panic(err)
	}
}
