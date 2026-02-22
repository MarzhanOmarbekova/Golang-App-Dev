package users

import (
	"Practice3/internal/repository/_postgres"
	"Practice3/pkg/modules"
	"fmt"
	"time"
)

type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var usersList []modules.User
	err := r.db.DB.Select(&usersList, "SELECT id, name, email, age, created_at FROM users")
	if err != nil {
		return nil, fmt.Errorf("GetUsers: %w", err)
	}
	return usersList, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	err := r.db.DB.Get(&user, `SELECT id, name, email, age, created_at FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: user with id=%d not found: %w", id, err)
	}
	return &user, nil
}

func (r *Repository) CreateUser(req modules.CreateUserRequest) (int, error) {
	if req.Name == "" {
		return 0, fmt.Errorf("CreateUser: name is required")
	}

	var newID int
	query := `INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.DB.QueryRow(query, req.Name, req.Email, req.Age).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return newID, nil
}

func (r *Repository) UpdateUser(id int, req modules.UpdateUserRequest) error {
	if req.Name == "" {
		return fmt.Errorf("UpdateUser: name is required")
	}

	query := `UPDATE users SET name=$1, email=$2, age=$3 WHERE id=$4`
	result, err := r.db.DB.Exec(query, req.Name, req.Email, req.Age, id)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateUser: failed to get rows affected: %w", err)
	}
	if rowAffected == 0 {
		return fmt.Errorf("UpdateUser: user with id=%d does not exist", id)
	}
	return nil
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	result, err := r.db.DB.Exec("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return 0, fmt.Errorf("DeleteUser: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("DeleteUser: failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("DeleteUser: user with id=%d does not exist", id)
	}
	return rowsAffected, nil
}
