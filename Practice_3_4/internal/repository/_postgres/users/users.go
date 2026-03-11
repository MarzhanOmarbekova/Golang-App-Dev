package users

import (
	"Practice3/internal/repository/_postgres"
	"Practice3/pkg/modules"
	"database/sql"
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

func softDeleteClause(status string) string {
	switch strings.ToLower(status) {
	case "deleted":
		return "deleted_at IS NOT NULL"
	case "all":
		return ""
	default: // "active" or blank
		return "deleted_at IS NULL"
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var list []modules.User
	err := r.db.DB.Select(&list,
		`SELECT id, name, email, age, gender, birth_date, created_at, deleted_at
		   FROM users
		  WHERE deleted_at IS NULL
		  ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("GetUsers: %w", err)
	}
	return list, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var u modules.User
	err := r.db.DB.Get(&u,
		`SELECT id, name, email, age, gender, birth_date, created_at, deleted_at
		   FROM users
		  WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: user with id=%d not found: %w", id, err)
	}
	return &u, nil
}

func (r *Repository) CreateUser(req modules.CreateUserRequest) (int, error) {
	if req.Name == "" {
		return 0, fmt.Errorf("CreateUser: name is required")
	}
	var newID int
	query := `INSERT INTO users (name, email, age, gender, birth_date)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.DB.QueryRow(query,
		req.Name, req.Email, req.Age, req.Gender, req.BirthDate,
	).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return newID, nil
}

func (r *Repository) UpdateUser(id int, req modules.UpdateUserRequest) error {
	if req.Name == "" {
		return fmt.Errorf("UpdateUser: name is required")
	}
	query := `UPDATE users
	             SET name=$1, email=$2, age=$3, gender=$4, birth_date=$5
	           WHERE id=$6 AND deleted_at IS NULL`
	result, err := r.db.DB.Exec(query,
		req.Name, req.Email, req.Age, req.Gender, req.BirthDate, id)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateUser: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("UpdateUser: user with id=%d does not exist or is already deleted", id)
	}
	return nil
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	result, err := r.db.DB.Exec(
		`UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return 0, fmt.Errorf("DeleteUser: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("DeleteUser: rows affected: %w", err)
	}
	if rows == 0 {
		return 0, fmt.Errorf("DeleteUser: user with id=%d does not exist or is already deleted", id)
	}
	return rows, nil
}

func (r *Repository) RestoreUser(id int) error {
	result, err := r.db.DB.Exec(
		`UPDATE users SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("RestoreUser: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("RestoreUser: rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("RestoreUser: user with id=%d not found or is not deleted", id)
	}
	return nil
}

func (r *Repository) GetPaginatedUsers(
	page, pageSize int,
	filter modules.UserFilter,
	sort modules.UserSort,
) (modules.PaginatedResponse, error) {

	args := []interface{}{}
	conditions := []string{}
	idx := 1

	if sc := softDeleteClause(filter.Status); sc != "" {
		conditions = append(conditions, sc)
	}

	if filter.ID != nil {
		conditions = append(conditions, fmt.Sprintf("id = $%d", idx))
		args = append(args, *filter.ID)
		idx++
	}
	if filter.Name != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", idx))
		args = append(args, "%"+*filter.Name+"%")
		idx++
	}
	if filter.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", idx))
		args = append(args, "%"+*filter.Email+"%")
		idx++
	}
	if filter.Gender != nil {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", idx))
		args = append(args, *filter.Gender)
		idx++
	}
	if filter.BirthDate != nil {
		conditions = append(conditions, fmt.Sprintf("birth_date = $%d", idx))
		args = append(args, *filter.BirthDate)
		idx++
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

	var totalCount int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	if err := r.db.DB.QueryRow(countSQL, args...).Scan(&totalCount); err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers count: %w", err)
	}

	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(
		`SELECT id, name, email, age, gender, birth_date, created_at, deleted_at
		   FROM users
		  %s
		  ORDER BY %s
		  LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, idx, idx+1,
	)
	dataArgs := append(args, pageSize, offset)

	rows, err := r.db.DB.Query(dataSQL, dataArgs...)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers query: %w", err)
	}
	defer rows.Close()

	var list []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Age,
			&u.Gender, &u.BirthDate, &u.CreatedAt, &u.DeletedAt,
		); err != nil {
			return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers scan: %w", err)
		}
		list = append(list, u)
	}

	return modules.PaginatedResponse{
		Data:       list,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) GetCursorPaginatedUsers(
	cursor, limit int,
	filter modules.UserFilter,
) (modules.CursorPaginatedResponse, error) {

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	args := []interface{}{}
	conditions := []string{}
	idx := 1

	if sc := softDeleteClause(filter.Status); sc != "" {
		conditions = append(conditions, sc)
	}

	if cursor > 0 {
		conditions = append(conditions, fmt.Sprintf("id > $%d", idx))
		args = append(args, cursor)
		idx++
	}

	if filter.Name != nil {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", idx))
		args = append(args, "%"+*filter.Name+"%")
		idx++
	}
	if filter.Gender != nil {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", idx))
		args = append(args, *filter.Gender)
		idx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(
		`SELECT id, name, email, age, gender, birth_date, created_at, deleted_at
		   FROM users
		  %s
		  ORDER BY id ASC
		  LIMIT $%d`,
		whereClause, idx,
	)
	args = append(args, limit+1)

	rows, err := r.db.DB.Query(query, args...)
	if err != nil {
		return modules.CursorPaginatedResponse{}, fmt.Errorf("GetCursorPaginatedUsers: %w", err)
	}
	defer rows.Close()

	var list []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Age,
			&u.Gender, &u.BirthDate, &u.CreatedAt, &u.DeletedAt,
		); err != nil {
			return modules.CursorPaginatedResponse{}, fmt.Errorf("GetCursorPaginatedUsers scan: %w", err)
		}
		list = append(list, u)
	}

	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}
	nextCursor := 0
	if hasMore && len(list) > 0 {
		nextCursor = list[len(list)-1].ID
	}

	return modules.CursorPaginatedResponse{
		Data:       list,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (r *Repository) GetCommonFriends(user1ID, user2ID int) ([]modules.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.age, u.gender, u.birth_date, u.created_at, u.deleted_at
		  FROM user_friends uf1
		  JOIN user_friends uf2 ON uf1.friend_id = uf2.friend_id
		  JOIN users u          ON u.id = uf1.friend_id
		 WHERE uf1.user_id = $1
		   AND uf2.user_id = $2
		   AND u.deleted_at IS NULL
		 ORDER BY u.id`

	rows, err := r.db.DB.Query(query, user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("GetCommonFriends: %w", err)
	}
	defer rows.Close()

	var result []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Age,
			&u.Gender, &u.BirthDate, &u.CreatedAt, &u.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("GetCommonFriends scan: %w", err)
		}
		result = append(result, u)
	}
	return result, nil
}

func (r *Repository) AddFriend(userID, friendID int) error {
	if userID == friendID {
		return fmt.Errorf("AddFriend: a user cannot be friends with themselves")
	}

	tx, err := r.db.DB.Begin()
	if err != nil {
		return fmt.Errorf("AddFriend: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback() //nolint:errcheck
		}
	}()

	insert := `INSERT INTO user_friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	if _, err = tx.Exec(insert, userID, friendID); err != nil {
		return fmt.Errorf("AddFriend: insert %d->%d: %w", userID, friendID, err)
	}
	if _, err = tx.Exec(insert, friendID, userID); err != nil {
		return fmt.Errorf("AddFriend: insert %d->%d: %w", friendID, userID, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("AddFriend: commit: %w", err)
	}
	return nil
}

func (r *Repository) GetFriendRecommendations(userID int) ([]modules.FriendRecommendation, error) {
	query := `
		SELECT
		    u.id,
		    u.name,
		    u.email,
		    u.age,
		    u.gender,
		    u.birth_date,
		    u.created_at,
		    u.deleted_at,
		    COUNT(uf1.friend_id) AS mutual_friends_count
		FROM user_friends uf1
		JOIN user_friends uf2
		    ON uf2.user_id = uf1.friend_id
		JOIN users u
		    ON u.id = uf2.friend_id
		LEFT JOIN user_friends uf_existing
		    ON uf_existing.user_id  = $1
		   AND uf_existing.friend_id = uf2.friend_id
		WHERE uf1.user_id = $1
		  AND uf2.friend_id <> $1
		  AND uf_existing.friend_id IS NULL
		  AND u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.email, u.age, u.gender, u.birth_date, u.created_at, u.deleted_at
		ORDER BY mutual_friends_count DESC, u.id ASC`

	rows, err := r.db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetFriendRecommendations: %w", err)
	}
	defer rows.Close()

	var results []modules.FriendRecommendation
	for rows.Next() {
		var u modules.User
		var mutualCount int
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.Age,
			&u.Gender, &u.BirthDate, &u.CreatedAt, &u.DeletedAt,
			&mutualCount,
		); err != nil {
			return nil, fmt.Errorf("GetFriendRecommendations scan: %w", err)
		}
		results = append(results, modules.FriendRecommendation{
			User:               u,
			MutualFriendsCount: mutualCount,
		})
	}
	return results, nil
}

var _ = sql.ErrNoRows
