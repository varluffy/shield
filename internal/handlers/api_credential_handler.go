// Package handlers provides HTTP request handlers.
// This file contains API credential handler for blacklist API key management.
package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/middleware"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// ApiCredentialHandler API密钥处理器
type ApiCredentialHandler struct {
	credentialService services.ApiCredentialService
	logger            *logger.Logger
	responseWriter    *response.ResponseWriter
}

// NewApiCredentialHandler 创建API密钥处理器
func NewApiCredentialHandler(
	credentialService services.ApiCredentialService,
	logger *logger.Logger,
) *ApiCredentialHandler {
	return &ApiCredentialHandler{
		credentialService: credentialService,
		logger:            logger,
		responseWriter:    response.NewResponseWriter(logger),
	}
}

// CreateApiCredential 创建API密钥
// @Summary 创建API密钥
// @Description 创建新的黑名单API密钥
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateApiCredentialRequest true "创建请求"
// @Success 201 {object} response.Response{data=dto.CreateApiCredentialResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials [post]
func (h *ApiCredentialHandler) CreateApiCredential(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CreateApiCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "参数绑定失败",
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrValidationFailed("参数绑定失败"))
		return
	}

	// 获取租户ID
	tenantID, exists := middleware.GetCurrentTenantID(c)
	if !exists {
		h.logger.ErrorWithTrace(ctx, "租户ID未找到")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	tenantIDUint64, _ := strconv.ParseUint(tenantID, 10, 64)

	// 转换为模型
	credential := &models.BlacklistApiCredential{
		TenantModel: models.TenantModel{TenantID: tenantIDUint64},
		Name:        req.Name,
		Description: req.Description,
		RateLimit:   req.RateLimit,
		IPWhitelist: req.IPWhitelist,
		ExpiresAt:   req.ExpiresAt,
	}

	// 创建API密钥
	apiKey, apiSecret, err := h.credentialService.CreateCredential(ctx, credential)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "创建API密钥失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.String("name", req.Name),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("创建失败"))
		return
	}

	// 构建响应（只在创建时返回完整密钥）
	resp := dto.CreateApiCredentialResponse{
		ApiCredentialInfo: dto.NewApiCredentialInfo(credential),
		APISecret:         apiSecret, // 仅在创建时返回
	}

	h.logger.InfoWithTrace(ctx, "创建API密钥成功",
		zap.Uint64("tenant_id", tenantIDUint64),
		zap.String("api_key", apiKey),
		zap.String("name", req.Name))

	h.responseWriter.Success(c, resp)
}

// GetApiCredentials 获取API密钥列表
// @Summary 获取API密钥列表
// @Description 获取租户的所有API密钥
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]dto.ApiCredentialInfo}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials [get]
func (h *ApiCredentialHandler) GetApiCredentials(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取租户ID
	tenantID, exists := middleware.GetCurrentTenantID(c)
	if !exists {
		h.logger.ErrorWithTrace(ctx, "租户ID未找到")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	tenantIDUint64, _ := strconv.ParseUint(tenantID, 10, 64)

	// 获取API密钥列表
	credentials, err := h.credentialService.GetCredentialsByTenant(ctx, tenantIDUint64)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "获取API密钥列表失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("获取列表失败"))
		return
	}

	// 构建响应
	items := make([]dto.ApiCredentialInfo, len(credentials))
	for i, credential := range credentials {
		items[i] = dto.NewApiCredentialInfo(credential)
	}

	h.responseWriter.Success(c, items)
}

// GetApiCredential 获取单个API密钥信息
// @Summary 获取API密钥详情
// @Description 获取指定的API密钥详细信息
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} response.Response{data=dto.ApiCredentialInfo}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials/{id} [get]
func (h *ApiCredentialHandler) GetApiCredential(c *gin.Context) {
	ctx := c.Request.Context()

	apiKey := c.Param("api_key")
	if apiKey == "" {
		h.logger.WarnWithTrace(ctx, "API Key参数为空")
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	// 获取API密钥信息
	credential, err := h.credentialService.GetCredentialByAPIKey(ctx, apiKey)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "获取API密钥信息失败",
			zap.String("api_key", apiKey),
			zap.Error(err))
		h.responseWriter.Error(c, errors.NewBusinessError(errors.CodeNotFound))
		return
	}

	// 验证权限（确保只能查看自己租户的密钥）
	tenantID, _ := middleware.GetCurrentTenantID(c)
	tenantIDUint64, _ := strconv.ParseUint(tenantID, 10, 64)
	if credential.TenantID != tenantIDUint64 {
		h.responseWriter.Error(c, errors.ErrForbidden())
		return
	}

	resp := dto.NewApiCredentialInfo(credential)
	h.responseWriter.Success(c, resp)
}

// UpdateApiCredential 更新API密钥
// @Summary 更新API密钥
// @Description 更新API密钥的配置信息
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Param request body dto.UpdateApiCredentialRequest true "更新请求"
// @Success 200 {object} response.Response{data=dto.ApiCredentialInfo}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials/{id} [put]
func (h *ApiCredentialHandler) UpdateApiCredential(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "ID参数格式错误",
			zap.String("id", idStr),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	var req dto.UpdateApiCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "参数绑定失败",
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrValidationFailed("参数绑定失败"))
		return
	}

	// 构建更新模型
	credential := &models.BlacklistApiCredential{
		TenantModel: models.TenantModel{
			ID: id,
		},
		Name:        req.Name,
		Description: req.Description,
		RateLimit:   req.RateLimit,
		IPWhitelist: req.IPWhitelist,
		ExpiresAt:   req.ExpiresAt,
	}

	// 更新API密钥
	err = h.credentialService.UpdateCredential(ctx, credential)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "更新API密钥失败",
			zap.Uint64("id", id),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("更新失败"))
		return
	}

	h.logger.InfoWithTrace(ctx, "更新API密钥成功",
		zap.Uint64("id", id))

	h.responseWriter.Success(c, nil)
}

// UpdateApiCredentialStatus 更新API密钥状态
// @Summary 更新API密钥状态
// @Description 启用/禁用/暂停API密钥
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Param request body dto.UpdateStatusRequest true "状态更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials/{id}/status [put]
func (h *ApiCredentialHandler) UpdateApiCredentialStatus(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "ID参数格式错误",
			zap.String("id", idStr),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "参数绑定失败",
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrValidationFailed("参数绑定失败"))
		return
	}

	// 更新状态
	err = h.credentialService.UpdateStatus(ctx, id, req.Status)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "更新API密钥状态失败",
			zap.Uint64("id", id),
			zap.String("status", req.Status),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("更新状态失败"))
		return
	}

	h.logger.InfoWithTrace(ctx, "更新API密钥状态成功",
		zap.Uint64("id", id),
		zap.String("status", req.Status))

	h.responseWriter.Success(c, nil)
}

// DeleteApiCredential 删除API密钥
// @Summary 删除API密钥
// @Description 删除指定的API密钥
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials/{id} [delete]
func (h *ApiCredentialHandler) DeleteApiCredential(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "ID参数格式错误",
			zap.String("id", idStr),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	// 删除API密钥
	err = h.credentialService.DeleteCredential(ctx, id)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "删除API密钥失败",
			zap.Uint64("id", id),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("删除失败"))
		return
	}

	h.logger.InfoWithTrace(ctx, "删除API密钥成功",
		zap.Uint64("id", id))

	h.responseWriter.Success(c, nil)
}

// RegenerateApiSecret 重新生成API Secret
// @Summary 重新生成API Secret
// @Description 重新生成API密钥的Secret
// @Tags API密钥管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API密钥ID"
// @Success 200 {object} response.Response{data=dto.RegenerateSecretResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/api-credentials/{id}/regenerate-secret [post]
func (h *ApiCredentialHandler) RegenerateApiSecret(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "ID参数格式错误",
			zap.String("id", idStr),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	// 重新生成Secret
	newSecret, err := h.credentialService.RegenerateSecret(ctx, id)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "重新生成API Secret失败",
			zap.Uint64("id", id),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("重新生成失败"))
		return
	}

	resp := dto.RegenerateSecretResponse{
		APISecret: newSecret,
		Message:   "API Secret已重新生成，请妥善保存，此密钥仅显示一次",
	}

	h.logger.InfoWithTrace(ctx, "重新生成API Secret成功",
		zap.Uint64("id", id))

	h.responseWriter.Success(c, resp)
}