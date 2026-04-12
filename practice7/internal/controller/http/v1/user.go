package v1

import (
	"net/http"
	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/pkg/logger"
	"practice-7/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type userRoutes struct {
	t   usecase.UserInterface
	l   logger.Interface
	rdb *redis.Client
}

func newUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface, rdb *redis.Client) {
	r := &userRoutes{t, l, rdb}

	rateLimiter := utils.NewRateLimiter(4, time.Minute)

	h := handler.Group("/users")
	h.Use(rateLimiter.Middleware())
	{
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)
		h.POST("/refresh", r.RefreshToken)
		h.POST("/verify-email", r.VerifyEmail)

		protected := h.Group("/")
		protected.Use(utils.JWTAuthMiddleware(rdb))
		{
			protected.GET("/me", r.GetMe)
			protected.GET("/protected/hello", r.ProtectedFunc)
			protected.POST("/logout", r.Logout)

			admin := protected.Group("/")
			admin.Use(utils.RoleMiddleware("admin"))
			{
				admin.PATCH("/promote/:id", r.PromoteUser)
			}
		}
	}
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
	var dto entity.CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(dto.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	role := dto.Role
	if role == "" {
		role = "user"
	}

	user := entity.User{
		ID:       uuid.New(),
		Username: dto.Username,
		Email:    dto.Email,
		Password: hashedPassword,
		Role:     role,
	}

	createdUser, sessionID, err := r.t.RegisterUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "User registered successfully. Please check your email for verification code.",
		"session_id": sessionID,
		"user":       createdUser,
	})
}

func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (r *userRoutes) GetMe(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := r.t.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"verified": user.Verified,
	})
}

func (r *userRoutes) ProtectedFunc(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"user_id": userID,
		"role":    role,
	})
}

func (r *userRoutes) PromoteUser(c *gin.Context) {
	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	updatedUser, err := r.t.PromoteUser(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User promoted to admin",
		"user":    updatedUser,
	})
}

func (r *userRoutes) VerifyEmail(c *gin.Context) {
	var dto entity.VerifyEmailDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.t.VerifyEmail(dto.Username, dto.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can now log in."})
}

func (r *userRoutes) RefreshToken(c *gin.Context) {
	var dto entity.RefreshTokenDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := r.t.RefreshToken(dto.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (r *userRoutes) Logout(c *gin.Context) {
	userIDStr, _ := c.Get("userID")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := r.t.LogoutUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
