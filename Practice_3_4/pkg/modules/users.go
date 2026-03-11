package modules

import "time"

type User struct {
	ID        int        `db:"id"         json:"id"`
	Name      string     `db:"name"       json:"name"`
	Email     string     `db:"email"      json:"email"`
	Age       int        `db:"age"        json:"age"`
	Gender    string     `db:"gender"     json:"gender"`
	BirthDate time.Time  `db:"birth_date" json:"birth_date"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"` // nil = active
}

type PaginatedResponse struct {
	Data       []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

type CursorPaginatedResponse struct {
	Data       []User `json:"data"`
	NextCursor int    `json:"next_cursor"` // ID of the last item; 0 means no more pages
	HasMore    bool   `json:"has_more"`
}

type UserFilter struct {
	ID        *int
	Name      *string
	Email     *string
	Gender    *string
	BirthDate *time.Time
	Status    string // "active" | "deleted" | "all"
}

type UserSort struct {
	Column    string // must be in the repository's allowedColumns whitelist
	Direction string // "ASC" or "DESC"
}

type FriendRecommendation struct {
	User               User `json:"user"`
	MutualFriendsCount int  `json:"mutual_friends_count"`
}

type AddFriendRequest struct {
	UserID   int `json:"user_id"`
	FriendID int `json:"friend_id"`
}

type CreateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
}

type UpdateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
}
