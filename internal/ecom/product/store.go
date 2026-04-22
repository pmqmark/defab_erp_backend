package product

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// ListProducts returns paginated products visible on web.
// Filters: categoryID, search (name ILIKE), minPrice, maxPrice, inStockOnly, attributes map.
// sortBy: "in_stock" (default), "newest", "price_asc", "price_desc".
func (s *Store) ListProducts(
	categoryID, search string,
	minPrice, maxPrice float64,
	inStockOnly bool,
	attributes map[string]string,
	sortBy string,
	page, limit int,
) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * limit

	where := "WHERE p.is_active = true AND p.is_web_visible = true"
	args := []interface{}{}
	idx := 1

	if categoryID != "" {
		where += " AND p.category_id = $" + strconv.Itoa(idx)
		args = append(args, categoryID)
		idx++
	}
	if search != "" {
		where += " AND (p.name ILIKE $" + strconv.Itoa(idx) + " OR p.brand ILIKE $" + strconv.Itoa(idx) + ")"
		args = append(args, "%"+search+"%")
		idx++
	}
	if minPrice > 0 {
		where += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM variants v WHERE v.product_id = p.id AND v.is_active = true AND v.price >= $%d
		)`, idx)
		args = append(args, minPrice)
		idx++
	}
	if maxPrice > 0 {
		where += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM variants v WHERE v.product_id = p.id AND v.is_active = true AND v.price <= $%d
		)`, idx)
		args = append(args, maxPrice)
		idx++
	}
	if inStockOnly {
		where += ` AND EXISTS (
			SELECT 1 FROM variants v
			JOIN stocks st ON st.variant_id = v.id
			JOIN warehouses w ON w.id = st.warehouse_id
			WHERE v.product_id = p.id AND v.is_active = true
			  AND w.type = 'CENTRAL' AND st.quantity > 0
		)`
	}
	// One correlated sub-select per attribute key/value pair
	for attrName, attrVal := range attributes {
		where += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM variants v2
			JOIN variant_attribute_mapping vam ON vam.variant_id = v2.id
			JOIN attribute_values av ON av.id = vam.attribute_value_id
			JOIN attributes a ON a.id = av.attribute_id
			WHERE v2.product_id = p.id AND v2.is_active = true
			  AND LOWER(a.name) = LOWER($%d) AND LOWER(av.value) = LOWER($%d)
		)`, idx, idx+1)
		args = append(args, attrName, attrVal)
		idx += 2
	}

	// Count (same WHERE, no ORDER/LIMIT)
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := s.db.QueryRow("SELECT COUNT(*) FROM products p "+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Order by
	var orderClause string
	switch sortBy {
	case "newest":
		orderClause = "ORDER BY p.created_at DESC"
	case "price_asc":
		orderClause = `ORDER BY
			(SELECT MIN(v.price) FROM variants v WHERE v.product_id = p.id AND v.is_active = true)
			ASC NULLS LAST`
	case "price_desc":
		orderClause = `ORDER BY
			(SELECT MAX(v.price) FROM variants v WHERE v.product_id = p.id AND v.is_active = true)
			DESC NULLS LAST`
	default: // "in_stock" — in-stock items bubble to top, then newest
		orderClause = `ORDER BY
			(EXISTS (
				SELECT 1 FROM variants v
				JOIN stocks st ON st.variant_id = v.id
				JOIN warehouses w ON w.id = st.warehouse_id
				WHERE v.product_id = p.id AND v.is_active = true
				  AND w.type = 'CENTRAL' AND st.quantity > 0
			)) DESC,
			p.created_at DESC`
	}

	query := `
		SELECT p.id, p.name, COALESCE(p.description, ''), COALESCE(p.brand, ''),
		       COALESCE(p.main_image_url, ''), COALESCE(c.name, '') AS category,
		       p.uom,
		       (SELECT MIN(v.price) FROM variants v WHERE v.product_id = p.id AND v.is_active = true) AS min_price,
		       (SELECT MAX(v.price) FROM variants v WHERE v.product_id = p.id AND v.is_active = true) AS max_price
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		` + where + `
		` + orderClause + `
		LIMIT $` + strconv.Itoa(idx) + ` OFFSET $` + strconv.Itoa(idx+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id, name, desc, brand, image, category, uom string
		var minP, maxP sql.NullFloat64

		if err := rows.Scan(&id, &name, &desc, &brand, &image, &category, &uom, &minP, &maxP); err != nil {
			return nil, 0, err
		}

		p := map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": desc,
			"brand":       brand,
			"image":       image,
			"category":    category,
			"uom":         uom,
		}
		if minP.Valid {
			p["min_price"] = minP.Float64
		}
		if maxP.Valid {
			p["max_price"] = maxP.Float64
		}
		products = append(products, p)
	}

	_ = int(math.Ceil(float64(total) / float64(limit)))
	return products, total, nil
}

// GetProductDetail returns a product with its variants (CENTRAL warehouse stock), images, and attributes.
func (s *Store) GetProductDetail(productID string) (map[string]interface{}, error) {
	var id, name, desc, brand, image, category, uom string
	var fabricComp, pattern, occasion, care sql.NullString

	err := s.db.QueryRow(`
		SELECT p.id, p.name, COALESCE(p.description, ''), COALESCE(p.brand, ''),
		       COALESCE(p.main_image_url, ''), COALESCE(c.name, '') AS category,
		       p.uom, p.fabric_composition, p.pattern, p.occasion, p.care_instructions
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1 AND p.is_active = true AND p.is_web_visible = true
	`, productID).Scan(&id, &name, &desc, &brand, &image, &category, &uom,
		&fabricComp, &pattern, &occasion, &care)
	if err != nil {
		return nil, err
	}

	product := map[string]interface{}{
		"id":                 id,
		"name":               name,
		"description":        desc,
		"brand":              brand,
		"image":              image,
		"category":           category,
		"uom":                uom,
		"fabric_composition": nullStr(fabricComp),
		"pattern":            nullStr(pattern),
		"occasion":           nullStr(occasion),
		"care_instructions":  nullStr(care),
	}

	// Product gallery images
	imgRows, err := s.db.Query(`SELECT image_url FROM product_images WHERE product_id = $1`, productID)
	if err == nil {
		defer imgRows.Close()
		var images []string
		for imgRows.Next() {
			var url string
			imgRows.Scan(&url)
			images = append(images, url)
		}
		product["images"] = images
	}

	// Variants — CENTRAL warehouse is the source of truth for ecom stock
	varRows, err := s.db.Query(`
		SELECT v.id, v.variant_code, v.name, v.sku, v.price, COALESCE(v.barcode, ''),
		       COALESCE(SUM(CASE WHEN w.type = 'CENTRAL' THEN st.quantity ELSE 0 END), 0) AS central_stock
		FROM variants v
		LEFT JOIN stocks st ON st.variant_id = v.id
		LEFT JOIN warehouses w ON w.id = st.warehouse_id
		WHERE v.product_id = $1 AND v.is_active = true
		GROUP BY v.id, v.variant_code, v.name, v.sku, v.price, v.barcode
		ORDER BY v.variant_code
	`, productID)
	if err != nil {
		return product, nil
	}
	defer varRows.Close()

	var variants []map[string]interface{}
	for varRows.Next() {
		var vid, vname, sku, barcode string
		var variantCode int
		var price, stock float64

		if err := varRows.Scan(&vid, &variantCode, &vname, &sku, &price, &barcode, &stock); err != nil {
			continue
		}

		v := map[string]interface{}{
			"id":           vid,
			"variant_code": variantCode,
			"name":         vname,
			"sku":          sku,
			"price":        price,
			"barcode":      barcode,
			"stock":        stock,
			"in_stock":     stock > 0,
		}

		// Variant images
		viRows, err := s.db.Query(`SELECT image_url FROM variant_images WHERE variant_id = $1`, vid)
		if err == nil {
			var vimages []string
			for viRows.Next() {
				var url string
				viRows.Scan(&url)
				vimages = append(vimages, url)
			}
			viRows.Close()
			if len(vimages) > 0 {
				v["images"] = vimages
			}
		}

		// Variant attributes (e.g. {"Color":"Red","Size":"M"})
		attrRows, err := s.db.Query(`
			SELECT a.name, av.value
			FROM variant_attribute_mapping vam
			JOIN attribute_values av ON av.id = vam.attribute_value_id
			JOIN attributes a ON a.id = av.attribute_id
			WHERE vam.variant_id = $1
		`, vid)
		if err == nil {
			attrs := map[string]string{}
			for attrRows.Next() {
				var aName, aVal string
				attrRows.Scan(&aName, &aVal)
				attrs[aName] = aVal
			}
			attrRows.Close()
			if len(attrs) > 0 {
				v["attributes"] = attrs
			}
		}

		variants = append(variants, v)
	}
	product["variants"] = variants

	return product, nil
}

// ListCategories returns active categories with live web-visible product counts.
func (s *Store) ListCategories() ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.name, COUNT(DISTINCT p.id) AS products_count
		FROM categories c
		LEFT JOIN products p ON p.category_id = c.id
		  AND p.is_active = true AND p.is_web_visible = true
		WHERE c.is_active = true
		GROUP BY c.id, c.name
		ORDER BY c.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []map[string]interface{}
	for rows.Next() {
		var id, name string
		var count int
		rows.Scan(&id, &name, &count)
		cats = append(cats, map[string]interface{}{
			"id":             id,
			"name":           name,
			"products_count": count,
		})
	}
	return cats, nil
}

// SearchSuggestions returns up to 8 product name / brand matches for autocomplete.
func (s *Store) SearchSuggestions(q string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT suggestion FROM (
			(SELECT DISTINCT p.name AS suggestion
			 FROM products p
			 WHERE p.is_active = true AND p.is_web_visible = true AND p.name ILIKE $1
			 LIMIT 8)
			UNION
			(SELECT DISTINCT p.brand AS suggestion
			 FROM products p
			 WHERE p.is_active = true AND p.is_web_visible = true
			   AND p.brand IS NOT NULL AND p.brand ILIKE $1
			 LIMIT 8)
		) sub
		ORDER BY suggestion
		LIMIT 8
	`, "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var sg string
		rows.Scan(&sg)
		suggestions = append(suggestions, sg)
	}
	return suggestions, nil
}

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
