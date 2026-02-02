package auth

import (
	"defab-erp/internal/core/model"

	"github.com/gofiber/fiber/v2"
)

// Injecting the Store interface is a good practice for testing
type AuthHandler struct {
	store *Store
}

func NewHandler(s *Store) *AuthHandler {
	return &AuthHandler{store: s}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// 1. Parse Input directly into Model (or a dedicated DTO if you prefer strict validation)
	u := new(model.User)
	if err := c.BodyParser(u); err != nil {
		return c.Status(400).SendString("Invalid input")
	}

	// 2. Call Store
	// Note: In real app, hash password here first
	if err := h.store.CreateUser(u); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(u)
}
