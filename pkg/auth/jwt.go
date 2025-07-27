// Package auth provides authentication and authorization utilities.
// It includes JWT token generation, validation, and user authentication helpers.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}

// JWTService JWT服务接口
type JWTService interface {
	GenerateAccessToken(userID, email, tenantID string) (string, error)
	GenerateRefreshToken(userID, tenantID string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshAccessToken(refreshToken string) (string, error)
}

// JWTServiceImpl JWT服务实现
type JWTServiceImpl struct {
	secretKey        string
	issuer           string
	accessTokenExp   time.Duration
	refreshTokenExp  time.Duration
}

// NewJWTService 创建JWT服务
func NewJWTService(secretKey, issuer string, accessTokenExp, refreshTokenExp time.Duration) JWTService {
	return &JWTServiceImpl{
		secretKey:       secretKey,
		issuer:          issuer,
		accessTokenExp:  accessTokenExp,
		refreshTokenExp: refreshTokenExp,
	}
}

// GenerateAccessToken 生成访问令牌
func (j *JWTServiceImpl) GenerateAccessToken(userID, email, tenantID string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		Email:    email,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenExp)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTServiceImpl) GenerateRefreshToken(userID, tenantID string) (string, error) {
	now := time.Now()
	claims := &JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExp)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken 验证令牌
func (j *JWTServiceImpl) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshAccessToken 刷新访问令牌
func (j *JWTServiceImpl) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// 检查是否为刷新令牌（刷新令牌通常不包含email和role）
	if claims.Email != "" || claims.TenantID != "" {
		return "", errors.New("not a refresh token")
	}

	// 这里应该从数据库获取最新的用户信息
	// 为了简化，我们只使用UserID生成新的访问令牌
	// 实际应用中应该查询数据库获取最新的用户信息
	return j.GenerateAccessToken(claims.UserID, "", claims.TenantID)
} 