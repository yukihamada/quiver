package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthConfig struct {
	JWTSecret     []byte
	APIKeyPrefix  string
	EnableAuth    bool
}

type Claims struct {
	UserID string `json:"user_id"`
	Plan   string `json:"plan"`
	jwt.RegisteredClaims
}

type Authenticator struct {
	config AuthConfig
}

func NewAuthenticator(config AuthConfig) *Authenticator {
	return &Authenticator{
		config: config,
	}
}

// AuthMiddleware validates JWT tokens or API keys
func (a *Authenticator) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.config.EnableAuth {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Check for Bearer token (JWT)
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := a.ValidateJWT(tokenString)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
				c.Abort()
				return
			}

			// Set user context
			c.Set("user_id", claims.UserID)
			c.Set("plan", claims.Plan)
			c.Next()
			return
		}

		// Check for API key
		if strings.HasPrefix(authHeader, a.config.APIKeyPrefix) {
			apiKey := authHeader
			userInfo, err := a.ValidateAPIKey(apiKey)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid API key",
				})
				c.Abort()
				return
			}

			// Set user context
			c.Set("user_id", userInfo.UserID)
			c.Set("plan", userInfo.Plan)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid authorization format",
		})
		c.Abort()
	}
}

// ValidateJWT validates a JWT token and returns claims
func (a *Authenticator) ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.config.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// UserInfo represents API key validation result
type UserInfo struct {
	UserID string
	Plan   string
}

// ValidateAPIKey validates an API key
func (a *Authenticator) ValidateAPIKey(apiKey string) (*UserInfo, error) {
	// Extract key parts
	parts := strings.Split(apiKey, "_")
	if len(parts) < 3 {
		return nil, jwt.ErrInvalidKey
	}

	// Verify HMAC signature
	keyData := strings.Join(parts[:len(parts)-1], "_")
	signature := parts[len(parts)-1]

	h := hmac.New(sha256.New, a.config.JWTSecret)
	h.Write([]byte(keyData))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, jwt.ErrSignatureInvalid
	}

	// TODO: Look up key in database
	// For now, return mock data
	return &UserInfo{
		UserID: "user_" + parts[1],
		Plan:   "pro",
	}, nil
}

// GenerateJWT generates a new JWT token
func (a *Authenticator) GenerateJWT(userID, plan string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Plan:   plan,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.config.JWTSecret)
}

// GenerateAPIKey generates a new API key
func (a *Authenticator) GenerateAPIKey(userID string) string {
	timestamp := time.Now().Unix()
	keyData := fmt.Sprintf("%s_%s_%d", a.config.APIKeyPrefix, userID, timestamp)
	
	h := hmac.New(sha256.New, a.config.JWTSecret)
	h.Write([]byte(keyData))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	
	return fmt.Sprintf("%s_%s", keyData, signature)
}