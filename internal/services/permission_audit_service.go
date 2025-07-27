// Package services contains business logic implementations.
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

//go:generate mockgen -source=permission_audit_service.go -destination=mocks/permission_audit_service_mock.go

// PermissionAuditService 权限审计服务接口
type PermissionAuditService interface {
	// LogPermissionGrant 记录权限授予操作
	LogPermissionGrant(ctx context.Context, req LogPermissionRequest) error
	// LogPermissionRevoke 记录权限撤销操作
	LogPermissionRevoke(ctx context.Context, req LogPermissionRequest) error
	// LogRoleCreate 记录角色创建操作
	LogRoleCreate(ctx context.Context, req LogRoleRequest) error
	// LogRoleUpdate 记录角色更新操作
	LogRoleUpdate(ctx context.Context, req LogRoleUpdateRequest) error
	// LogRoleDelete 记录角色删除操作
	LogRoleDelete(ctx context.Context, req LogRoleRequest) error
	// LogUserRoleAssign 记录用户角色分配操作
	LogUserRoleAssign(ctx context.Context, req LogUserRoleRequest) error
	// LogUserRoleRevoke 记录用户角色撤销操作
	LogUserRoleRevoke(ctx context.Context, req LogUserRoleRequest) error
	// GetAuditLogs 获取审计日志
	GetAuditLogs(ctx context.Context, filter repositories.AuditLogFilter) ([]models.PermissionAuditLog, int64, error)
}

// LogPermissionRequest 权限操作日志请求
type LogPermissionRequest struct {
	TenantID       uint64 `json:"tenant_id"`
	OperatorID     uint64 `json:"operator_id"`
	TargetType     string `json:"target_type"` // role, user
	TargetID       uint64 `json:"target_id"`
	PermissionCode string `json:"permission_code"`
	Reason         string `json:"reason"`
	IPAddress      string `json:"ip_address"`
	UserAgent      string `json:"user_agent"`
}

// LogRoleRequest 角色操作日志请求
type LogRoleRequest struct {
	TenantID   uint64 `json:"tenant_id"`
	OperatorID uint64 `json:"operator_id"`
	RoleID     uint64 `json:"role_id"`
	RoleData   string `json:"role_data"`
	Reason     string `json:"reason"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}

// LogRoleUpdateRequest 角色更新日志请求
type LogRoleUpdateRequest struct {
	LogRoleRequest
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// LogUserRoleRequest 用户角色操作日志请求
type LogUserRoleRequest struct {
	TenantID   uint64 `json:"tenant_id"`
	OperatorID uint64 `json:"operator_id"`
	UserID     uint64 `json:"user_id"`
	RoleID     uint64 `json:"role_id"`
	RoleName   string `json:"role_name"`
	Reason     string `json:"reason"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}

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

// permissionAuditService 权限审计服务实现
type permissionAuditService struct {
	auditRepo repositories.PermissionAuditRepository
	logger    *logger.Logger
}

// NewPermissionAuditService 创建权限审计服务
func NewPermissionAuditService(
	auditRepo repositories.PermissionAuditRepository,
	logger *logger.Logger,
) PermissionAuditService {
	return &permissionAuditService{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// LogPermissionGrant 记录权限授予操作
func (s *permissionAuditService) LogPermissionGrant(ctx context.Context, req LogPermissionRequest) error {
	return s.logPermissionOperation(ctx, req, models.AuditActionGrant)
}

// LogPermissionRevoke 记录权限撤销操作
func (s *permissionAuditService) LogPermissionRevoke(ctx context.Context, req LogPermissionRequest) error {
	return s.logPermissionOperation(ctx, req, models.AuditActionRevoke)
}

// LogRoleCreate 记录角色创建操作
func (s *permissionAuditService) LogRoleCreate(ctx context.Context, req LogRoleRequest) error {
	auditLog := &models.PermissionAuditLog{
		TenantID:    req.TenantID,
		OperatorID:  req.OperatorID,
		TargetType:  models.AuditTargetRole,
		TargetID:    req.RoleID,
		Action:      models.AuditActionCreate,
		NewValue:    req.RoleData,
		Reason:      req.Reason,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
	}

	return s.createAuditLog(ctx, auditLog)
}

// LogRoleUpdate 记录角色更新操作
func (s *permissionAuditService) LogRoleUpdate(ctx context.Context, req LogRoleUpdateRequest) error {
	auditLog := &models.PermissionAuditLog{
		TenantID:    req.TenantID,
		OperatorID:  req.OperatorID,
		TargetType:  models.AuditTargetRole,
		TargetID:    req.RoleID,
		Action:      models.AuditActionUpdate,
		OldValue:    req.OldValue,
		NewValue:    req.NewValue,
		Reason:      req.Reason,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
	}

	return s.createAuditLog(ctx, auditLog)
}

// LogRoleDelete 记录角色删除操作
func (s *permissionAuditService) LogRoleDelete(ctx context.Context, req LogRoleRequest) error {
	auditLog := &models.PermissionAuditLog{
		TenantID:    req.TenantID,
		OperatorID:  req.OperatorID,
		TargetType:  models.AuditTargetRole,
		TargetID:    req.RoleID,
		Action:      models.AuditActionDelete,
		OldValue:    req.RoleData,
		Reason:      req.Reason,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
	}

	return s.createAuditLog(ctx, auditLog)
}

// LogUserRoleAssign 记录用户角色分配操作
func (s *permissionAuditService) LogUserRoleAssign(ctx context.Context, req LogUserRoleRequest) error {
	return s.logUserRoleOperation(ctx, req, models.AuditActionGrant)
}

// LogUserRoleRevoke 记录用户角色撤销操作
func (s *permissionAuditService) LogUserRoleRevoke(ctx context.Context, req LogUserRoleRequest) error {
	return s.logUserRoleOperation(ctx, req, models.AuditActionRevoke)
}

// GetAuditLogs 获取审计日志
func (s *permissionAuditService) GetAuditLogs(ctx context.Context, filter repositories.AuditLogFilter) ([]models.PermissionAuditLog, int64, error) {
	s.logger.DebugWithTrace(ctx, "Getting audit logs",
		zap.Any("filter", filter))

	// 设置默认分页参数
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	logs, total, err := s.auditRepo.List(ctx, filter)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get audit logs",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved audit logs",
		zap.Int("count", len(logs)),
		zap.Int64("total", total))

	return logs, total, nil
}

// logPermissionOperation 记录权限操作
func (s *permissionAuditService) logPermissionOperation(ctx context.Context, req LogPermissionRequest, action string) error {
	auditLog := &models.PermissionAuditLog{
		TenantID:       req.TenantID,
		OperatorID:     req.OperatorID,
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		Action:         action,
		PermissionCode: req.PermissionCode,
		Reason:         req.Reason,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
	}

	return s.createAuditLog(ctx, auditLog)
}

// logUserRoleOperation 记录用户角色操作
func (s *permissionAuditService) logUserRoleOperation(ctx context.Context, req LogUserRoleRequest, action string) error {
	roleData := map[string]interface{}{
		"role_id":   req.RoleID,
		"role_name": req.RoleName,
	}
	
	roleDataJSON, _ := json.Marshal(roleData)

	auditLog := &models.PermissionAuditLog{
		TenantID:   req.TenantID,
		OperatorID: req.OperatorID,
		TargetType: models.AuditTargetUser,
		TargetID:   req.UserID,
		Action:     action,
		NewValue:   string(roleDataJSON),
		Reason:     req.Reason,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	return s.createAuditLog(ctx, auditLog)
}

// createAuditLog 创建审计日志
func (s *permissionAuditService) createAuditLog(ctx context.Context, auditLog *models.PermissionAuditLog) error {
	err := s.auditRepo.Create(ctx, auditLog)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to create audit log",
			zap.String("action", auditLog.Action),
			zap.String("target_type", auditLog.TargetType),
			zap.Uint64("target_id", auditLog.TargetID),
			zap.Error(err))
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Audit log created",
		zap.String("action", auditLog.Action),
		zap.String("target_type", auditLog.TargetType),
		zap.Uint64("target_id", auditLog.TargetID),
		zap.Uint64("operator_id", auditLog.OperatorID))

	return nil
}