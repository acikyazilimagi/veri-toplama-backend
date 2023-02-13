package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Middlewares(a *fiber.App) {
	a.Use(
		cors.New(),
		logger.New(),
		recover.New(),
		pprof.New(),
	)
}
