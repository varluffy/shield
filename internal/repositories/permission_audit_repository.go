// Package repositories contains data access layer implementations.
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

// AuditLogFilter 审计日志过滤器
type AuditLogFilter struct {
	TenantID   *uint64 `json:"tenant_id"`
	OperatorID *uint64 `json:"operator_id"`
	TargetType *string `json:"target_type"`
	TargetID   *uint64 `json:"target_id"`
	Action     *string `json:"action"`
	StartTime  *string `json:"start_time"`
	EndTime    *string `json:"end_time"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
}

//go:generate mockgen -source=permission_audit_repository.go -destination=mocks/permission_audit_repository_mock.go

// PermissionAuditRepository 权限审计仓储接口
type PermissionAuditRepository interface {
	Create(ctx context.Context, auditLog *models.PermissionAuditLog) error
	List(ctx context.Context, filter AuditLogFilter) ([]models.PermissionAuditLog, int64, error)
	GetByID(ctx context.Context, id uint64) (*models.PermissionAuditLog, error)
}

// PermissionAuditRepositoryImpl 权限审计仓储实现
type PermissionAuditRepositoryImpl struct {
	*transaction.BaseRepository
	logger *logger.Logger
}

// NewPermissionAuditRepository 创建权限审计仓储
func NewPermissionAuditRepository(db *gorm.DB, txManager transaction.TransactionManager, logger *logger.Logger) PermissionAuditRepository {
	return &PermissionAuditRepositoryImpl{
		BaseRepository: transaction.NewBaseRepository(db, txManager, logger.Logger),
		logger:         logger,
	}
}

// Create 创建审计日志
func (r *PermissionAuditRepositoryImpl) Create(ctx context.Context, auditLog *models.PermissionAuditLog) error {
	r.LogTransactionState(ctx, "Create Permission Audit Log")

	db := r.GetDB(ctx)
	err := db.Create(auditLog).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to create audit log",
			zap.String("action", auditLog.Action),
			zap.String("target_type", auditLog.TargetType),
			zap.Uint64("target_id", auditLog.TargetID),
			zap.Error(err))
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Created audit log",
		zap.Uint64("id", auditLog.ID),
		zap.String("action", auditLog.Action),
		zap.String("target_type", auditLog.TargetType))

	return nil
}

// List 获取审计日志列表
func (r *PermissionAuditRepositoryImpl) List(ctx context.Context, filter AuditLogFilter) ([]models.PermissionAuditLog, int64, error) {
	r.LogTransactionState(ctx, "List Permission Audit Logs")

	db := r.GetDB(ctx)
	var logs []models.PermissionAuditLog
	var total int64

	// 构建查询条件
	query := db.Model(&models.PermissionAuditLog{})

	if filter.TenantID != nil {
		query = query.Where("tenant_id = ?", *filter.TenantID)
	}

	if filter.OperatorID != nil {
		query = query.Where("operator_id = ?", *filter.OperatorID)
	}

	if filter.TargetType != nil {
		query = query.Where("target_type = ?", *filter.TargetType)
	}

	if filter.TargetID != nil {
		query = query.Where("target_id = ?", *filter.TargetID)
	}

	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to count audit logs",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.Limit
	err = query.Offset(offset).Limit(filter.Limit).
		Order("created_at DESC").
		Find(&logs).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to list audit logs",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved audit logs",
		zap.Int("count", len(logs)),
		zap.Int64("total", total))

	return logs, total, nil
}

// GetByID 根据ID获取审计日志
func (r *PermissionAuditRepositoryImpl) GetByID(ctx context.Context, id uint64) (*models.PermissionAuditLog, error) {
	r.LogTransactionState(ctx, "Get Audit Log By ID")

	db := r.GetDB(ctx)
	var auditLog models.PermissionAuditLog

	err := db.Where("id = ?", id).First(&auditLog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.DebugWithTrace(ctx, "Audit log not found", zap.Uint64("id", id))
			return nil, fmt.Errorf("audit log not found with id: %d", id)
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get audit log by ID",
			zap.Uint64("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved audit log by ID",
		zap.Uint64("id", id),
		zap.String("action", auditLog.Action))

	return &auditLog, nil
}
