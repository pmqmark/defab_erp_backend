package model

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	InvoiceNumber string    `gorm:"unique;not null;index" json:"invoice_number"`
	BranchID      uint      `json:"branch_id"`
	TotalAmount   float64   `json:"total_amount"`
	TaxAmount     float64   `json:"tax_amount"`
	Type          string    `json:"type"` // "POS" or "WEB"
	CreatedAt     time.Time `json:"created_at"`

	Payments []Payment `gorm:"foreignKey:InvoiceID" json:"payments"`
}

type Payment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	InvoiceID uuid.UUID `json:"invoice_id"`
	Method    string    `json:"method"` // "Cash", "Card", "UPI"
	Amount    float64   `json:"amount"`
	Reference string    `json:"reference"`
}
