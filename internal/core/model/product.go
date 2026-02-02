package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"not null;index" json:"name"` // Index for search
	Description string    `json:"description"`
	CategoryID  uint      `json:"category_id"`
	Category    Category  `json:"category"`
	Brand       string    `json:"brand"`

	// --- BUSINESS LOGIC FLAGS ---
	IsWebVisible bool   `gorm:"default:true;index" json:"is_web_visible"` // Index for fast Web Catalog filtering
	IsStitched   bool   `gorm:"default:false" json:"is_stitched"`         // True = Shirt (Unit), False = Fabric (Meter)
	UOM          string `gorm:"default:'Unit'" json:"uom"`                // "Unit" or "Meter"

	Variants []Variant `gorm:"foreignKey:ProductID" json:"variants"`
}

type Variant struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`                             // e.g., "Size M", "Red Roll"
	SKU       string    `gorm:"unique;not null;index" json:"sku"` // Unique Index is critical
	Price     float64   `gorm:"not null" json:"price"`
	CostPrice float64   `json:"cost_price"`

	Barcodes []Barcode `gorm:"foreignKey:VariantID" json:"barcodes"`
}

type Barcode struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	VariantID   uuid.UUID `json:"variant_id"`
	Code        string    `gorm:"unique;not null;index" json:"code"` // Index for instant POS scanning
	GeneratedAt time.Time `json:"generated_at"`
}
