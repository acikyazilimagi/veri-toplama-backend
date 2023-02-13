package app

import (
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const APP_NAME = "veri-toplama-backend"

type App struct {
	App *fiber.App
}

func NewApp() *App {
	return &App{
		App: fiber.New(fiber.Config{
			AppName:       APP_NAME,
			ReadTimeout:   time.Second * time.Duration(30),
			WriteTimeout:  time.Second * time.Duration(30),
			CaseSensitive: true,
			BodyLimit:     64 * 1024 * 1024,
			Concurrency:   256 * 1024,
			IdleTimeout:   10 * time.Second,
		}),
	}
}

func (a *App) Run() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("application gracefully shutting down..")
		a.App.Shutdown()
	}()

	if err := a.App.Listen(":80"); err != nil {
		panic(fmt.Sprintf("app error: %s", err.Error()))
	}
}

func (a *App) Register() {
	a.App.Get("/healthcheck", handler.Healtcheck)
	a.App.Get("/monitor", monitor.New())
	// TODO: add to swagger
}
