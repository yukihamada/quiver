package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/quiver/billing/pkg/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID   uuid.UUID        `json:"user_id"`
	Email    string           `json:"email"`
	Plan     models.PlanType  `json:"plan"`
	APIKey   string           `json:"api_key,omitempty"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT operations
type TokenManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewTokenManager creates a new token manager
func NewTokenManager(secretKey string, tokenDuration time.Duration) *TokenManager {
	return &TokenManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken creates a JWT token for a user
func (tm *TokenManager) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Plan:   user.Plan,
		APIKey: user.APIKey,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "quiver-billing",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// ValidateToken validates and parses a JWT token
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64 URL-safe string
	apiKey := base64.URLEncoding.EncodeToString(bytes)
	
	// Add prefix for easy identification
	return "sk_live_" + apiKey, nil
}

// ValidateAPIKey checks if an API key format is valid
func ValidateAPIKey(apiKey string) bool {
	// Check prefix and length
	if len(apiKey) < 50 || apiKey[:8] != "sk_live_" {
		return false
	}
	
	// Try to decode the base64 part
	_, err := base64.URLEncoding.DecodeString(apiKey[8:])
	return err == nil
}

// HashAPIKey creates a hash of the API key for storage
func HashAPIKey(apiKey string) string {
	// In production, use proper hashing like bcrypt
	// For now, return as-is (TODO: implement proper hashing)
	return apiKey
}