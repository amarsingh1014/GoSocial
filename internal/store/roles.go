package store

import (
	"context"
	"database/sql"
	"log"
)

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type RolesStore struct {
	db *sql.DB
}

func (r *RolesStore) GetByName(ctx context.Context, name string) (*Role, error) {
	query := `SELECT id, name, level, description FROM roles WHERE name = $1`

	role := &Role{}

	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name,
		&role.Level, &role.Description)

	if err != nil {
		log.Printf("here")
		return nil, err
	}

	return role, nil
}
