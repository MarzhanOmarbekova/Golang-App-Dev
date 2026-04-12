package entity

type CreateUserDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // Optional: "user" (default) or "admin"
}

type LoginUserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailDTO struct {
	Username string `json:"username" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
