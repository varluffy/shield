package models

import (
	"time"
)

// RefreshToken 刷新令牌模型（不需要UUID）
type RefreshToken struct {
	BaseModelWithoutUUID
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	TenantID  uint64    `gorm:"not null;index" json:"tenant_id"`
	TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_hash"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}