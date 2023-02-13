package handler

import (
	"github.com/YusufOzmen01/veri-kontrol-backend/services"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type servicesClient interface {
	GeetFeeds(c *fiber.Ctx) (*services.FeetFeedsResponse, error)
}

func GetFeeds(servicesClient servicesClient) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		res, err := servicesClient.GeetFeeds(ctx)
		if err != nil {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": err,
			})
		}

		return ctx.Status(http.StatusOK).JSON(res)
	}
}
