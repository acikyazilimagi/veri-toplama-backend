package handler

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type resolveServiceClient interface {
	Resolve(c *fiber.Ctx) error
}

func Resolve(resolveServiceClient resolveServiceClient) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := resolveServiceClient.Resolve(ctx); err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err,
			})
		}

		return ctx.SendStatus(http.StatusCreated)
	}
}
