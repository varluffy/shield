// Package repositories provides data access layer implementations.
// This file contains API credential repository for blacklist API key management.
package repositories

import (
	"context"
	"time"

	"github.com/varluffy/shield/internal/models"
	"gorm.io/gorm"
)

// ApiCredentialRepository API密钥仓储接口
type ApiCredentialRepository interface {
	Create(ctx context.Context, credential *models.BlacklistApiCredential) error
	GetByAPIKey(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error)
	GetByTenant(ctx context.Context, tenantID uint64) ([]*models.BlacklistApiCredential, error)
	Update(ctx context.Context, credential *models.BlacklistApiCredential) error
	UpdateLastUsedAt(ctx context.Context, apiKey string) error
	Delete(ctx context.Context, id uint64) error
	GetActiveByAPIKey(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error)
}

// apiCredentialRepository API密钥仓储实现
type apiCredentialRepository struct {
	db *gorm.DB
}

// NewApiCredentialRepository 创建API密钥仓储
func NewApiCredentialRepository(db *gorm.DB) ApiCredentialRepository {
	return &apiCredentialRepository{
		db: db,
	}
}

// Create 创建API密钥记录
func (r *apiCredentialRepository) Create(ctx context.Context, credential *models.BlacklistApiCredential) error {
	return r.db.WithContext(ctx).Create(credential).Error
}

// GetByAPIKey 根据API Key获取密钥记录
func (r *apiCredentialRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error) {
	var credential models.BlacklistApiCredential
	err := r.db.WithContext(ctx).
		Where("api_key = ?", apiKey).
		First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}

// GetByTenant 获取租户的所有API密钥
func (r *apiCredentialRepository) GetByTenant(ctx context.Context, tenantID uint64) ([]*models.BlacklistApiCredential, error) {
	var credentials []*models.BlacklistApiCredential
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&credentials).Error
	return credentials, err
}

// Update 更新API密钥记录
func (r *apiCredentialRepository) Update(ctx context.Context, credential *models.BlacklistApiCredential) error {
	return r.db.WithContext(ctx).Save(credential).Error
}

// UpdateLastUsedAt 更新最后使用时间
func (r *apiCredentialRepository) UpdateLastUsedAt(ctx context.Context, apiKey string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.BlacklistApiCredential{}).
		Where("api_key = ?", apiKey).
		Update("last_used_at", now).Error
}

// Delete 删除API密钥记录（软删除）
func (r *apiCredentialRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.BlacklistApiCredential{}, id).Error
}

// GetActiveByAPIKey 获取有效的API密钥记录（状态检查+过期检查）
func (r *apiCredentialRepository) GetActiveByAPIKey(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error) {
	var credential models.BlacklistApiCredential
	query := r.db.WithContext(ctx).Where("api_key = ? AND status = ?", apiKey, "active")

	// 如果设置了过期时间，检查是否过期
	query = query.Where("(expires_at IS NULL OR expires_at > ?)", time.Now())

	err := query.First(&credential).Error
	if err != nil {
		return nil, err
	}
	return &credential, nil
}
