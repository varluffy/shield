// Package services provides business logic layer implementations.
// This file contains API credential service for blacklist API key management.
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

// ApiCredentialService API密钥服务接口
type ApiCredentialService interface {
	CreateCredential(ctx context.Context, credential *models.BlacklistApiCredential) (apiKey, apiSecret string, err error)
	GetCredentialByAPIKey(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error)
	GetCredentialsByTenant(ctx context.Context, tenantID uint64) ([]*models.BlacklistApiCredential, error)
	UpdateCredential(ctx context.Context, credential *models.BlacklistApiCredential) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
	DeleteCredential(ctx context.Context, id uint64) error
	RegenerateSecret(ctx context.Context, id uint64) (newSecret string, err error)
}

// apiCredentialService API密钥服务实现
type apiCredentialService struct {
	credentialRepo repositories.ApiCredentialRepository
	logger         *logger.Logger
}

// NewApiCredentialService 创建API密钥服务
func NewApiCredentialService(
	credentialRepo repositories.ApiCredentialRepository,
	logger *logger.Logger,
) ApiCredentialService {
	return &apiCredentialService{
		credentialRepo: credentialRepo,
		logger:         logger,
	}
}

// CreateCredential 创建API密钥
func (s *apiCredentialService) CreateCredential(ctx context.Context, credential *models.BlacklistApiCredential) (apiKey, apiSecret string, err error) {
	// 生成API Key (ak_前缀 + 32位随机字符)
	apiKey = "ak_" + generateRandomString(32)
	credential.APIKey = apiKey

	// 生成API Secret (64位随机字符)
	apiSecret = generateRandomString(64)
	credential.APISecret = apiSecret

	// 设置默认状态
	if credential.Status == "" {
		credential.Status = "active"
	}

	// 设置默认限流值
	if credential.RateLimit == 0 {
		credential.RateLimit = 1000
	}

	// 创建记录
	err = s.credentialRepo.Create(ctx, credential)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "创建API密钥失败",
			zap.Uint64("tenant_id", credential.TenantID),
			zap.String("name", credential.Name),
			zap.Error(err))
		return "", "", fmt.Errorf("创建API密钥失败: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "API密钥创建成功",
		zap.Uint64("tenant_id", credential.TenantID),
		zap.String("api_key", apiKey),
		zap.String("name", credential.Name))

	// 返回生成的密钥（仅在创建时返回一次）
	return apiKey, apiSecret, nil
}

// GetCredentialByAPIKey 根据API Key获取密钥信息
func (s *apiCredentialService) GetCredentialByAPIKey(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error) {
	credential, err := s.credentialRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		s.logger.WarnWithTrace(ctx, "获取API密钥信息失败",
			zap.String("api_key", apiKey),
			zap.Error(err))
		return nil, err
	}

	// 不返回secret给客户端
	credential.APISecret = ""
	return credential, nil
}

// GetCredentialsByTenant 获取租户的所有API密钥
func (s *apiCredentialService) GetCredentialsByTenant(ctx context.Context, tenantID uint64) ([]*models.BlacklistApiCredential, error) {
	credentials, err := s.credentialRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "获取租户API密钥列表失败",
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err))
		return nil, err
	}

	// 清除所有secret
	for _, credential := range credentials {
		credential.APISecret = ""
	}

	return credentials, nil
}

// UpdateCredential 更新API密钥信息
func (s *apiCredentialService) UpdateCredential(ctx context.Context, credential *models.BlacklistApiCredential) error {
	// 不允许更新APIKey和APISecret
	existingCredential, err := s.credentialRepo.GetByAPIKey(ctx, credential.APIKey)
	if err != nil {
		return fmt.Errorf("获取现有密钥信息失败: %w", err)
	}

	// 保留原有的APIKey和APISecret
	credential.APIKey = existingCredential.APIKey
	credential.APISecret = existingCredential.APISecret

	err = s.credentialRepo.Update(ctx, credential)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "更新API密钥失败",
			zap.Uint64("id", credential.ID),
			zap.String("api_key", credential.APIKey),
			zap.Error(err))
		return fmt.Errorf("更新API密钥失败: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "API密钥更新成功",
		zap.Uint64("id", credential.ID),
		zap.String("api_key", credential.APIKey))

	return nil
}

// UpdateStatus 更新API密钥状态
func (s *apiCredentialService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	// 验证状态值
	validStatuses := map[string]bool{
		"active":    true,
		"inactive":  true,
		"suspended": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("无效的状态值: %s", status)
	}

	// 获取现有记录
	var credential models.BlacklistApiCredential
	credential.ID = id
	credential.Status = status

	err := s.credentialRepo.Update(ctx, &credential)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "更新API密钥状态失败",
			zap.Uint64("id", id),
			zap.String("status", status),
			zap.Error(err))
		return fmt.Errorf("更新API密钥状态失败: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "API密钥状态更新成功",
		zap.Uint64("id", id),
		zap.String("status", status))

	return nil
}

// DeleteCredential 删除API密钥
func (s *apiCredentialService) DeleteCredential(ctx context.Context, id uint64) error {
	err := s.credentialRepo.Delete(ctx, id)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "删除API密钥失败",
			zap.Uint64("id", id),
			zap.Error(err))
		return fmt.Errorf("删除API密钥失败: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "API密钥删除成功",
		zap.Uint64("id", id))

	return nil
}

// RegenerateSecret 重新生成API密钥的Secret
func (s *apiCredentialService) RegenerateSecret(ctx context.Context, id uint64) (newSecret string, err error) {
	// 生成新的Secret
	newSecret = generateRandomString(64)

	// 更新记录
	credential := &models.BlacklistApiCredential{
		TenantModel: models.TenantModel{
			ID: id,
		},
		APISecret: newSecret,
	}

	err = s.credentialRepo.Update(ctx, credential)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "重新生成API Secret失败",
			zap.Uint64("id", id),
			zap.Error(err))
		return "", fmt.Errorf("重新生成API Secret失败: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "API Secret重新生成成功",
		zap.Uint64("id", id))

	return newSecret, nil
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}