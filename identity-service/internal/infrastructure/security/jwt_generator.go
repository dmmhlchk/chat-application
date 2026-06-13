package security

import (
	"errors"
	"fmt"
	"identity-service/internal/application/port"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var _ port.TokenGenerator = (*JWTGenerator)(nil)

// struct for payload
type UserClaims struct {
	UserID    uuid.UUID `json:"sub"`
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

type JWTGenerator struct {
	secretKey []byte
}

func NewJWTGenerator(secret string) *JWTGenerator {
	return &JWTGenerator{secretKey: []byte(secret)}
}

func (g *JWTGenerator) GenerateToken(userID uuid.UUID, sessionID uuid.UUID, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := UserClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Create token with HMAC-SHA256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(g.secretKey)
	if err != nil {
		return "", fmt.Errorf("token signing failed: %w", err)
	}

	return signedToken, nil
}

func (g *JWTGenerator) ValidateToken(tokenString string) (uuid.UUID, uuid.UUID, error) {
	// Parse the token string with our target claims destination structure
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Crucial Security Check: Validate that the signing method matches what we expect
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected token signing algorithm: %v", t.Header["alg"])
		}
		return g.secretKey, nil
	})

	if err != nil {
		// Automatically handles expired tokens, invalid formats, and corrupted signatures
		return uuid.Nil, uuid.Nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Extract claims and verify the token status flag is fully valid
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return uuid.Nil, uuid.Nil, errors.New("token contains invalid claims infrastructure")
	}

	return claims.UserID, claims.SessionID, nil
}
