package product

import (
	"database/sql"
	"defab-erp/internal/core/model" // <--- Importing shared models
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Transaction logic for complex Product creation
func (s *Store) CreateProduct(p *model.Product) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Product
	err = tx.QueryRow(`
		INSERT INTO products (name, category_id, is_web_visible, is_stitched, uom)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		p.Name, p.CategoryID, p.IsWebVisible, p.IsStitched, p.UOM,
	).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("insert product failed: %w", err)
	}

	// 2. Insert Variants (Loop through the slice in the model)
	stmt, _ := tx.Prepare(`
		INSERT INTO variants (product_id, name, sku, price, cost_price)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`)
	defer stmt.Close()

	for i := range p.Variants {
		v := &p.Variants[i] // Use pointer to update ID back into the struct
		err = stmt.QueryRow(p.ID, v.Name, v.SKU, v.Price, v.CostPrice).Scan(&v.ID)
		if err != nil {
			return fmt.Errorf("insert variant failed: %w", err)
		}
	}

	return tx.Commit()
}
