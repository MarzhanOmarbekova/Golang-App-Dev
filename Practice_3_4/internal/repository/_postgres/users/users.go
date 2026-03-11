package users

import (
	"Practice3/internal/repository/_postgres"
	"Practice3/pkg/modules"
	"fmt"
	"strings"
	"time"
)

var allowedColumns = map[string]bool{
	"id":         true,
	"name":       true,
	"email":      true,
	"gender":     true,
	"birth_date": true,
	"age":        true,
	"created_at": true,
}

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

func (r *Repository) GetPaginatedUsers(
	page, pageSize int,
	filter modules.UserFilter,
	sort modules.UserSort,
) (modules.PaginatedResponse, error) {

	args := []interface{}{}
	conditions := []string{}
	argIdx := 1

	if filter.ID != nil {
		conditions = append(conditions, fmt.Sprintf("id = $%d", argIdx))
		args = append(args, *filter.ID)
		argIdx++
	}
	if filter.Name != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Name+"%")
		argIdx++
	}
	if filter.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Email+"%")
		argIdx++
	}
	if filter.Gender != nil {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argIdx))
		args = append(args, *filter.Gender)
		argIdx++
	}
	if filter.BirthDate != nil {
		conditions = append(conditions, fmt.Sprintf("birth_date = $%d", argIdx))
		args = append(args, *filter.BirthDate)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := "id ASC"
	if sort.Column != "" {
		if !allowedColumns[sort.Column] {
			return modules.PaginatedResponse{}, fmt.Errorf("invalid sort column: %s", sort.Column)
		}
		dir := "ASC"
		if strings.ToUpper(sort.Direction) == "DESC" {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", sort.Column, dir)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var totalCount int
	if err := r.db.DB.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers count: %w", err)
	}

	offset := (page - 1) * pageSize

	limitArg := argIdx
	offsetArg := argIdx + 1

	dataQuery := fmt.Sprintf(
		`SELECT id, name, email, age, gender, birth_date, created_at
		   FROM users
		  %s
		  ORDER BY %s
		  LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, limitArg, offsetArg,
	)
	dataArgs := append(args, pageSize, offset)

	rows, err := r.db.DB.Query(dataQuery, dataArgs...)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers query: %w", err)
	}
	defer rows.Close()

	var usersList []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Age,
			&u.Gender, &u.BirthDate, &u.CreatedAt,
		); err != nil {
			return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers scan: %w", err)
		}
		usersList = append(usersList, u)
	}

	return modules.PaginatedResponse{
		Data:       usersList,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) GetCommonFriends(user1ID, user2ID int) ([]modules.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.age, u.gender, u.birth_date, u.created_at
		  FROM user_friends uf1
		  JOIN user_friends uf2 ON uf1.friend_id = uf2.friend_id
		  JOIN users u          ON u.id = uf1.friend_id
		 WHERE uf1.user_id = $1
		   AND uf2.user_id = $2
		 ORDER BY u.id`

	rows, err := r.db.DB.Query(query, user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("GetCommonFriends query: %w", err)
	}
	defer rows.Close()

	var result []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Age,
			&u.Gender, &u.BirthDate, &u.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetCommonFriends scan: %w", err)
		}
		result = append(result, u)
	}
	return result, nil
}