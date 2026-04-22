package product

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers public (no auth) product catalog routes.
// /suggest and /categories must be registered before /:id so Fiber doesn't
// capture the literal strings as a product ID.
func RegisterRoutes(r fiber.Router, h *Handler) {
	r.Get("/suggest",    h.SearchSuggestions)
	r.Get("/categories", h.Categories)
	r.Get("/",           h.List)
	r.Get("/:id",        h.GetByID)
}
