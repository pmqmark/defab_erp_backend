package cart

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(r fiber.Router, h *Handler) {
	r.Get("/count",       h.Count)
	r.Get("/",            h.Get)
	r.Post("/items",      h.AddItem)
	r.Patch("/items/:id", h.UpdateItem)
	r.Delete("/items/:id", h.RemoveItem)
	r.Delete("/",         h.Clear)
}
