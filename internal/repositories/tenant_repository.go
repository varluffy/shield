// Package repositories contains data access layer implementations.
// It provides repository pattern implementations for database operations.
package repositories

import (
	"context"
	"fmt"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/transaction"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate mockgen -source=tenant_repository.go -destination=mocks/tenant_repository_mock.go

// TenantRepository 租户仓储接口
type TenantRepository interface {
	GetByID(ctx context.Context, id uint64) (*models.Tenant, error)
	GetByUUID(ctx context.Context, uuid string) (*models.Tenant, error)
	GetUUIDByID(ctx context.Context, id uint64) (string, error)
}

// TenantRepositoryImpl 租户仓储实现
type TenantRepositoryImpl struct {
	*transaction.BaseRepository
	logger *logger.Logger
}

// NewTenantRepository 创建租户仓储
func NewTenantRepository(db *gorm.DB, txManager transaction.TransactionManager, logger *logger.Logger) TenantRepository {
	return &TenantRepositoryImpl{
		BaseRepository: transaction.NewBaseRepository(db, txManager, logger.Logger),
		logger:         logger,
	}
}

// GetByID 根据ID获取租户
func (r *TenantRepositoryImpl) GetByID(ctx context.Context, id uint64) (*models.Tenant, error) {
	r.LogTransactionState(ctx, "Get Tenant By ID")
	
	var tenant models.Tenant
	db := r.GetDB(ctx)
	
	err := db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.DebugWithTrace(ctx, "Tenant not found", zap.Uint64("id", id))
			return nil, fmt.Errorf("tenant not found with id: %d", id)
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get tenant by ID",
			zap.Uint64("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	r.logger.DebugWithTrace(ctx, "Retrieved tenant by ID",
		zap.Uint64("id", id),
		zap.String("name", tenant.Name))
	
	return &tenant, nil
}

// GetByUUID 根据UUID获取租户
func (r *TenantRepositoryImpl) GetByUUID(ctx context.Context, uuid string) (*models.Tenant, error) {
	r.LogTransactionState(ctx, "Get Tenant By UUID")
	
	var tenant models.Tenant
	db := r.GetDB(ctx)
	
	err := db.Where("uuid = ?", uuid).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.DebugWithTrace(ctx, "Tenant not found", zap.String("uuid", uuid))
			return nil, fmt.Errorf("tenant not found with uuid: %s", uuid)
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get tenant by UUID",
			zap.String("uuid", uuid),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	r.logger.DebugWithTrace(ctx, "Retrieved tenant by UUID",
		zap.String("uuid", uuid),
		zap.String("name", tenant.Name))
	
	return &tenant, nil
}

// GetUUIDByID 根据ID获取租户UUID
func (r *TenantRepositoryImpl) GetUUIDByID(ctx context.Context, id uint64) (string, error) {
	r.LogTransactionState(ctx, "Get Tenant UUID By ID")
	
	var tenant models.Tenant
	db := r.GetDB(ctx)
	
	err := db.Select("uuid").Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.DebugWithTrace(ctx, "Tenant not found", zap.Uint64("id", id))
			return "", fmt.Errorf("tenant not found with id: %d", id)
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get tenant UUID by ID",
			zap.Uint64("id", id),
			zap.Error(err))
		return "", fmt.Errorf("failed to get tenant UUID: %w", err)
	}
	
	r.logger.DebugWithTrace(ctx, "Retrieved tenant UUID by ID",
		zap.Uint64("id", id),
		zap.String("uuid", tenant.UUID))
	
	return tenant.UUID, nil
}