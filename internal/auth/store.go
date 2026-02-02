package auth

import (
	"database/sql"
	"defab-erp/internal/core/model" // <--- Importing shared models
	"errors"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(u *model.User) error {
	// Standard SQL using the model fields
	query := `
		INSERT INTO users (name, email, password_hash, role_id, branch_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(query, u.Name, u.Email, u.PasswordHash, u.RoleID, u.BranchID).Scan(
		&u.ID, &u.CreatedAt,
	)
	return err
}

func (s *Store) GetUserByEmail(email string) (*model.User, error) {
	u := &model.User{} // Use shared model
	var roleName string

	query := `
		SELECT u.id, u.name, u.email, u.password_hash, u.role_id, r.name, u.branch_id, u.created_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`
	err := s.db.QueryRow(query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.RoleID, &roleName, &u.BranchID, &u.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	u.Role.Name = roleName // Populate nested struct
	return u, err
}
