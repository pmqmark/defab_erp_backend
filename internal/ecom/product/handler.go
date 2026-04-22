package product

import (
	"math"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	store *Store
}

func NewHandler(s *Store) *Handler {
	return &Handler{store: s}
}

// GET /api/ecom/products
// Query params: page, limit, category_id, q, min_price, max_price, in_stock, attr[Color]=Red, sort
func (h *Handler) List(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	categoryID := c.Query("category_id")
	search := c.Query("q")
	minPrice := c.QueryFloat("min_price", 0)
	maxPrice := c.QueryFloat("max_price", 0)
	inStockOnly := c.QueryBool("in_stock", false)
	sortBy := c.Query("sort", "in_stock")

	// Parse attribute filters: ?attr[Color]=Red&attr[Size]=M
	attributes := map[string]string{}
	c.Context().QueryArgs().VisitAll(func(key, val []byte) {
		k := string(key)
		// Fiber keeps bracket params as "attr[Color]"
		if len(k) > 5 && k[:5] == "attr[" && k[len(k)-1] == ']' {
			attrName := k[5 : len(k)-1]
			attributes[attrName] = string(val)
		}
	})

	products, total, err := h.store.ListProducts(
		categoryID, search,
		minPrice, maxPrice,
		inStockOnly,
		attributes,
		sortBy,
		page, limit,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch products"})
	}
	if products == nil {
		products = []map[string]interface{}{}
	}

	return c.JSON(fiber.Map{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		"data":        products,
	})
}

// GET /api/ecom/products/:id
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	product, err := h.store.GetProductDetail(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "product not found"})
	}
	return c.JSON(product)
}

// GET /api/ecom/products/categories
func (h *Handler) Categories(c *fiber.Ctx) error {
	cats, err := h.store.ListCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch categories"})
	}
	if cats == nil {
		cats = []map[string]interface{}{}
	}
	return c.JSON(fiber.Map{"categories": cats})
}

// GET /api/ecom/products/suggest?q=sari
func (h *Handler) SearchSuggestions(c *fiber.Ctx) error {
	q := c.Query("q")
	if len(q) < 2 {
		return c.JSON(fiber.Map{"suggestions": []string{}})
	}
	suggestions, err := h.store.SearchSuggestions(q)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch suggestions"})
	}
	if suggestions == nil {
		suggestions = []string{}
	}
	return c.JSON(fiber.Map{"suggestions": suggestions})
}