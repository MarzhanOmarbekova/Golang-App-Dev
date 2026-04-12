package utils

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	gomail "gopkg.in/gomail.v2"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateJWT(userID uuid.UUID, role string) (TokenPair, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	accessTTL, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	if accessTTL == 0 {
		accessTTL = 15
	}
	refreshTTL, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))
	if refreshTTL == 0 {
		refreshTTL = 10080
	}

	// Access token
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"type":    "access",
		"exp":     time.Now().Add(time.Minute * time.Duration(accessTTL)).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessSigned, err := accessToken.SignedString(secret)
	if err != nil {
		return TokenPair{}, err
	}

	// Refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"type":    "refresh",
		"exp":     time.Now().Add(time.Minute * time.Duration(refreshTTL)).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshSigned, err := refreshToken.SignedString(secret)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessSigned, RefreshToken: refreshSigned}, nil
}

func ParseJWT(tokenStr string) (jwt.MapClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func JWTAuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims, err := ParseJWT(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check token type
		if tokenType, ok := claims["type"].(string); ok && tokenType == "refresh" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Use access token, not refresh token"})
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Redis check: is this token still active?
		if rdb != nil {
			ctx := context.Background()
			stored, err := rdb.Get(ctx, "auth:"+userID).Result()
			if err != nil || stored != tokenStr {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is logged out or expired"})
				return
			}
		}

		c.Set("userID", userID)
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No role found"})
			return
		}
		if role.(string) != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Access denied. Required role: %s, your role: %s", requiredRole, role),
			})
			return
		}
		c.Next()
	}
}

type visitorInfo struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorInfo
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorInfo),
		limit:    limit,
		window:   window,
	}
	// Cleanup old entries every minute
	go func() {
		for range time.Tick(time.Minute) {
			rl.mu.Lock()
			for key, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.window {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			key = "user:" + userID.(string)
		} else {
			tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
			if tokenStr != "" {
				if claims, err := ParseJWT(tokenStr); err == nil {
					if uid, ok := claims["user_id"].(string); ok {
						key = "user:" + uid
					}
				}
			} else {
				key = "ip:" + c.ClientIP()
			}
		}

		rl.mu.Lock()
		v, exists := rl.visitors[key]
		if !exists || time.Since(v.lastSeen) > rl.window {
			rl.visitors[key] = &visitorInfo{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			c.Next()
			return
		}
		v.count++
		v.lastSeen = time.Now()
		if v.count > rl.limit {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("Rate limit exceeded. Max %d requests per %v.", rl.limit, rl.window),
			})
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}

func GenerateVerifyCode() string {
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func SendVerificationEmail(toEmail, code string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" {
		fmt.Printf("[EMAIL] Would send code %s to %s\n", code, toEmail)
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Your Verification Code")
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>Email Verification</h2>
		<p>Your verification code is: <strong>%s</strong></p>
		<p>This code expires in 10 minutes.</p>
	`, code))

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}
