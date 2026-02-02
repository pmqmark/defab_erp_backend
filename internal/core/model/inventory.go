package model

import (
	"time"

	"github.com/google/uuid"
)

type Branch struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Address   string    `json:"address"`
	ManagerID uuid.UUID `json:"manager_id"`
}

type Warehouse struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	BranchID *uint  `json:"branch_id"`
	Name     string `gorm:"not null" json:"name"`
	Type     string `gorm:"default:'STORE'" json:"type"` // "CENTRAL", "STORE", "FACTORY"
}

type Supplier struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"not null" json:"name"`
	Contact string `json:"contact"`
	Email   string `json:"email"`
}

type WarehouseStock struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	WarehouseID uint      `gorm:"index" json:"warehouse_id"`                    // Index for fast lookup
	VariantID   uuid.UUID `gorm:"index" json:"variant_id"`                      // Index for fast lookup
	Quantity    float64   `gorm:"type:decimal(10,2);default:0" json:"quantity"` // Decimal support for 1.5 meters
	Reserved    float64   `gorm:"type:decimal(10,2);default:0" json:"reserved"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StockMovement struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	VariantID       uuid.UUID `gorm:"index" json:"variant_id"`
	FromWarehouseID *uint     `json:"from_warehouse_id"`
	ToWarehouseID   *uint     `json:"to_warehouse_id"`
	Quantity        float64   `json:"quantity"`
	Type            string    `json:"type"` // "PURCHASE", "SALE", "TRANSFER", "MANUFACTURING"
	ReferenceID     string    `json:"reference_id"`
	CreatedAt       time.Time `json:"created_at"`
}
