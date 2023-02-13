package handler

import (
	"context"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	usersRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/users"
	"github.com/gofiber/fiber/v2"
	"os"
)

// TODO: Repodan çekiyor
func AdminHandler(c *fiber.Ctx) error {
	authKey := c.Get("Auth-Key")

	ctx := context.TODO()
	mongoURL := os.Getenv("MONGO_URL")
	mongoClient := sources.NewMongoClient(ctx, mongoURL, "database")
	userRepository := usersRepository.NewRepository(mongoClient)

	user, err := userRepository.GetUser(ctx, authKey)
	if err != nil {
		return c.Status(401).SendString("User not found.")
	}

	if user.PermLevel < usersRepository.PermModerator {
		return c.Status(401).SendString("You are not allowed to access here.")
	}

	return c.Next()
}
