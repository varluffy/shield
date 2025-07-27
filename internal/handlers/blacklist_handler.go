// Package handlers provides HTTP request handlers.
// This file contains blacklist handler for phone blacklist operations.
package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/middleware"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// BlacklistHandler 黑名单处理器
type BlacklistHandler struct {
	blacklistService services.BlacklistService
	logger           *logger.Logger
	responseWriter   *response.ResponseWriter
}

// NewBlacklistHandler 创建黑名单处理器
func NewBlacklistHandler(
	blacklistService services.BlacklistService,
	logger *logger.Logger,
) *BlacklistHandler {
	return &BlacklistHandler{
		blacklistService: blacklistService,
		logger:           logger,
		responseWriter:   response.NewResponseWriter(logger),
	}
}

// CheckBlacklist 检查手机号MD5是否在黑名单中
// @Summary 检查黑名单
// @Description 检查手机号MD5是否在黑名单中
// @Tags 黑名单查询
// @Accept json
// @Produce json
// @Param X-API-Key header string true "API密钥"
// @Param X-Timestamp header string true "时间戳"
// @Param X-Nonce header string true "随机数"
// @Param X-Signature header string true "HMAC签名"
// @Param request body dto.CheckBlacklistRequest true "查询请求"
// @Success 200 {object} response.Response{data=dto.CheckBlacklistResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 429 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/blacklist/check [post]
func (h *BlacklistHandler) CheckBlacklist(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()

	var req dto.CheckBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "参数绑定失败",
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrValidationFailed("参数绑定失败"))
		return
	}

	// 从上下文获取租户ID
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.ErrorWithTrace(ctx, "租户ID未找到")
		h.responseWriter.Error(c, errors.ErrInternalError("租户信息丢失"))
		return
	}

	tenantIDUint64, ok := tenantID.(uint64)
	if !ok {
		h.logger.ErrorWithTrace(ctx, "租户ID类型错误")
		h.responseWriter.Error(c, errors.ErrInternalError("租户信息格式错误"))
		return
	}

	// 设置上下文信息供日志中间件使用
	c.Set("phone_md5", req.PhoneMD5)

	// 检查黑名单
	isBlacklist, err := h.blacklistService.CheckPhoneMD5(ctx, tenantIDUint64, req.PhoneMD5)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "黑名单查询失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.String("phone_md5", req.PhoneMD5),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("查询失败"))
		return
	}

	// 设置结果供日志中间件使用
	c.Set("blacklist_result", isBlacklist)

	resp := dto.CheckBlacklistResponse{
		IsBlacklist: isBlacklist,
		PhoneMD5:    req.PhoneMD5,
	}

	h.logger.DebugWithTrace(ctx, "黑名单查询成功",
		zap.Uint64("tenant_id", tenantIDUint64),
		zap.String("phone_md5", req.PhoneMD5),
		zap.Bool("is_blacklist", isBlacklist),
		zap.Duration("duration", time.Since(start)))

	h.responseWriter.Success(c, resp)
}

// CreateBlacklist 创建黑名单记录
// @Summary 创建黑名单
// @Description 创建新的黑名单记录
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBlacklistRequest true "创建请求"
// @Success 201 {object} response.Response{data=dto.BlacklistInfo}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/blacklist [post]
func (h *BlacklistHandler) CreateBlacklist(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CreateBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "参数绑定失败",
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrValidationFailed("参数绑定失败"))
		return
	}

	// 获取当前用户信息
	userID, _, tenantID, exists := middleware.GetCurrentUser(c)
	if !exists {
		h.logger.ErrorWithTrace(ctx, "用户信息未找到")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	tenantIDUint64, _ := strconv.ParseUint(tenantID, 10, 64)
	operatorIDUint64, _ := strconv.ParseUint(userID, 10, 64)

	// 转换为模型
	blacklist := req.ToModel(tenantIDUint64, operatorIDUint64)

	// 创建黑名单记录
	err := h.blacklistService.CreateBlacklist(ctx, blacklist)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "创建黑名单记录失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.String("phone_md5", req.PhoneMD5),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("创建失败"))
		return
	}

	resp := dto.NewBlacklistInfo(blacklist)

	h.logger.InfoWithTrace(ctx, "创建黑名单记录成功",
		zap.Uint64("tenant_id", tenantIDUint64),
		zap.String("phone_md5", req.PhoneMD5),
		zap.Uint64("operator_id", operatorIDUint64))

	h.responseWriter.Success(c, resp)
}

// BatchImportBlacklist 批量导入黑名单
// @Summary 批量导入黑名单
// @Description 批量导入黑名单记录
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BatchImportBlacklistRequest true "批量导入请求"
// @Success 200 {object} response.Response{data=dto.BatchImportResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/blacklist/import [post]
func (h *BlacklistHandler) BatchImportBlacklist(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.BatchImportBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "参数绑定失败",
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrValidationFailed("参数绑定失败"))
		return
	}

	// 获取当前用户信息
	userID, _, tenantID, exists := middleware.GetCurrentUser(c)
	if !exists {
		h.logger.ErrorWithTrace(ctx, "用户信息未找到")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	tenantIDUint64, _ := strconv.ParseUint(tenantID, 10, 64)
	operatorIDUint64, _ := strconv.ParseUint(userID, 10, 64)

	// 批量导入
	err := h.blacklistService.BatchImportBlacklist(
		ctx,
		tenantIDUint64,
		req.PhoneMD5List,
		req.Source,
		req.Reason,
		operatorIDUint64,
	)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "批量导入黑名单失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.Int("count", len(req.PhoneMD5List)),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("导入失败"))
		return
	}

	resp := dto.BatchImportResponse{
		SuccessCount: len(req.PhoneMD5List),
		FailedCount:  0,
		FailedItems:  []string{},
	}

	h.logger.InfoWithTrace(ctx, "批量导入黑名单成功",
		zap.Uint64("tenant_id", tenantIDUint64),
		zap.Int("count", len(req.PhoneMD5List)),
		zap.Uint64("operator_id", operatorIDUint64))

	h.responseWriter.Success(c, resp)
}

// GetBlacklistList 获取黑名单列表
// @Summary 获取黑名单列表
// @Description 分页获取黑名单列表
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dto.GetBlacklistResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/blacklist [get]
func (h *BlacklistHandler) GetBlacklistList(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.GetBlacklistRequest
	if err := c.ShouldBindQuery(&req); err != nil {
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

	// 获取黑名单列表
	blacklists, total, err := h.blacklistService.GetBlacklistByTenant(
		ctx,
		tenantIDUint64,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "获取黑名单列表失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("获取列表失败"))
		return
	}

	// 构建响应
	items := make([]dto.BlacklistInfo, len(blacklists))
	for i, blacklist := range blacklists {
		items[i] = dto.NewBlacklistInfo(blacklist)
	}

	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)

	resp := dto.GetBlacklistResponse{
		Items: items,
		Pagination: dto.PaginationInfo{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	h.responseWriter.Success(c, resp)
}

// DeleteBlacklist 删除黑名单记录
// @Summary 删除黑名单
// @Description 删除指定的黑名单记录
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "黑名单ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/blacklist/{id} [delete]
func (h *BlacklistHandler) DeleteBlacklist(c *gin.Context) {
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

	// 删除黑名单记录
	err = h.blacklistService.DeleteBlacklist(ctx, id)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "删除黑名单记录失败",
			zap.Uint64("id", id),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("删除失败"))
		return
	}

	h.logger.InfoWithTrace(ctx, "删除黑名单记录成功",
		zap.Uint64("id", id))

	h.responseWriter.Success(c, nil)
}

// GetQueryStats 获取查询统计
// @Summary 获取查询统计
// @Description 获取黑名单查询统计信息
// @Tags 黑名单管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param hours query int false "统计小时数" default(24)
// @Success 200 {object} response.Response{data=dto.QueryStatsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/admin/blacklist/stats [get]
func (h *BlacklistHandler) GetQueryStats(c *gin.Context) {
	ctx := c.Request.Context()

	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 || hours > 168 { // 最多7天
		hours = 24
	}

	// 获取租户ID
	tenantID, exists := middleware.GetCurrentTenantID(c)
	if !exists {
		h.logger.ErrorWithTrace(ctx, "租户ID未找到")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	tenantIDUint64, _ := strconv.ParseUint(tenantID, 10, 64)

	// 获取统计信息
	stats, err := h.blacklistService.GetQueryStats(ctx, tenantIDUint64, hours)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "获取查询统计失败",
			zap.Uint64("tenant_id", tenantIDUint64),
			zap.Int("hours", hours),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("获取统计失败"))
		return
	}

	resp := dto.QueryStatsResponse{
		TotalQueries: stats.TotalQueries,
		HitCount:     stats.HitCount,
		MissCount:    stats.MissCount,
		HitRate:      stats.HitRate,
		AvgLatency:   stats.AvgLatency,
	}

	h.responseWriter.Success(c, resp)
}