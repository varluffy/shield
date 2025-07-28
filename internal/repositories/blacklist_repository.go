// Package repositories provides data access layer implementations.
// This file contains blacklist repository for phone blacklist operations.
package repositories

import (
	"context"

	"github.com/varluffy/shield/internal/models"
	"gorm.io/gorm"
)

// BlacklistRepository 黑名单仓储接口
type BlacklistRepository interface {
	Create(ctx context.Context, blacklist *models.PhoneBlacklist) error
	GetByID(ctx context.Context, id uint64) (*models.PhoneBlacklist, error)
	GetByTenantAndMD5(ctx context.Context, tenantID uint64, phoneMD5 string) (*models.PhoneBlacklist, error)
	GetByTenant(ctx context.Context, tenantID uint64, offset, limit int) ([]*models.PhoneBlacklist, int64, error)
	Update(ctx context.Context, blacklist *models.PhoneBlacklist) error
	Delete(ctx context.Context, id uint64) error
	BatchCreate(ctx context.Context, blacklists []*models.PhoneBlacklist) error
	GetActiveMD5ListByTenant(ctx context.Context, tenantID uint64) ([]string, error)
	ExistsByTenantAndMD5(ctx context.Context, tenantID uint64, phoneMD5 string) (bool, error)
}

// blacklistRepository 黑名单仓储实现
type blacklistRepository struct {
	db *gorm.DB
}

// NewBlacklistRepository 创建黑名单仓储
func NewBlacklistRepository(db *gorm.DB) BlacklistRepository {
	return &blacklistRepository{
		db: db,
	}
}

// Create 创建黑名单记录
func (r *blacklistRepository) Create(ctx context.Context, blacklist *models.PhoneBlacklist) error {
	return r.db.WithContext(ctx).Create(blacklist).Error
}

// GetByID 根据ID获取黑名单记录
func (r *blacklistRepository) GetByID(ctx context.Context, id uint64) (*models.PhoneBlacklist, error) {
	var blacklist models.PhoneBlacklist
	err := r.db.WithContext(ctx).First(&blacklist, id).Error
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

// GetByTenantAndMD5 根据租户ID和手机号MD5获取黑名单记录
func (r *blacklistRepository) GetByTenantAndMD5(ctx context.Context, tenantID uint64, phoneMD5 string) (*models.PhoneBlacklist, error) {
	var blacklist models.PhoneBlacklist
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND phone_md5 = ? AND is_active = ?", tenantID, phoneMD5, true).
		First(&blacklist).Error
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

// GetByTenant 分页获取租户的黑名单记录
func (r *blacklistRepository) GetByTenant(ctx context.Context, tenantID uint64, offset, limit int) ([]*models.PhoneBlacklist, int64, error) {
	var blacklists []*models.PhoneBlacklist
	var total int64

	// 获取总数
	err := r.db.WithContext(ctx).Model(&models.PhoneBlacklist{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = r.db.WithContext(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&blacklists).Error
	if err != nil {
		return nil, 0, err
	}

	return blacklists, total, nil
}

// Update 更新黑名单记录
func (r *blacklistRepository) Update(ctx context.Context, blacklist *models.PhoneBlacklist) error {
	return r.db.WithContext(ctx).Save(blacklist).Error
}

// Delete 删除黑名单记录（软删除）
func (r *blacklistRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.PhoneBlacklist{}, id).Error
}

// BatchCreate 批量创建黑名单记录
func (r *blacklistRepository) BatchCreate(ctx context.Context, blacklists []*models.PhoneBlacklist) error {
	if len(blacklists) == 0 {
		return nil
	}

	// 使用事务批量插入，提高性能
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(blacklists, 1000).Error
	})
}

// GetActiveMD5ListByTenant 获取租户所有有效的手机号MD5列表（用于Redis同步）
func (r *blacklistRepository) GetActiveMD5ListByTenant(ctx context.Context, tenantID uint64) ([]string, error) {
	var md5List []string
	err := r.db.WithContext(ctx).Model(&models.PhoneBlacklist{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Pluck("phone_md5", &md5List).Error
	return md5List, err
}

// ExistsByTenantAndMD5 检查指定租户和MD5是否存在黑名单记录
func (r *blacklistRepository) ExistsByTenantAndMD5(ctx context.Context, tenantID uint64, phoneMD5 string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PhoneBlacklist{}).
		Where("tenant_id = ? AND phone_md5 = ? AND is_active = ?", tenantID, phoneMD5, true).
		Count(&count).Error
	return count > 0, err
}
