// Package dto provides data transfer objects for API requests and responses.
// This file contains blacklist related DTOs.
package dto

import (
	"github.com/varluffy/shield/internal/models"
	"time"
)

// CheckBlacklistRequest 黑名单查询请求
type CheckBlacklistRequest struct {
	PhoneMD5 string `json:"phone_md5" binding:"required,len=32" example:"5d41402abc4b2a76b9719d911017c592"`
}

// CheckBlacklistResponse 黑名单查询响应
type CheckBlacklistResponse struct {
	IsBlacklist bool   `json:"is_blacklist" example:"true"`
	PhoneMD5    string `json:"phone_md5" example:"5d41402abc4b2a76b9719d911017c592"`
}

// CreateBlacklistRequest 创建黑名单请求
type CreateBlacklistRequest struct {
	PhoneMD5 string `json:"phone_md5" binding:"required,len=32" example:"5d41402abc4b2a76b9719d911017c592"`
	Source   string `json:"source" binding:"required" example:"manual"`
	Reason   string `json:"reason" example:"用户投诉"`
}

// ToModel 转换为模型
func (r *CreateBlacklistRequest) ToModel(tenantID, operatorID uint64) *models.PhoneBlacklist {
	return &models.PhoneBlacklist{
		TenantModel: models.TenantModel{TenantID: tenantID},
		PhoneMD5:    r.PhoneMD5,
		Source:      r.Source,
		Reason:      r.Reason,
		OperatorID:  operatorID,
		IsActive:    true,
	}
}

// BatchImportBlacklistRequest 批量导入黑名单请求
type BatchImportBlacklistRequest struct {
	PhoneMD5List []string `json:"phone_md5_list" binding:"required,min=1,max=10000"`
	Source       string   `json:"source" binding:"required" example:"import"`
	Reason       string   `json:"reason" example:"批量导入"`
}

// GetBlacklistRequest 获取黑名单列表请求
type GetBlacklistRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

// BlacklistInfo 黑名单信息
type BlacklistInfo struct {
	ID         uint64    `json:"id" example:"1"`
	UUID       string    `json:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	PhoneMD5   string    `json:"phone_md5" example:"5d41402abc4b2a76b9719d911017c592"`
	Source     string    `json:"source" example:"manual"`
	Reason     string    `json:"reason" example:"用户投诉"`
	OperatorID uint64    `json:"operator_id" example:"1"`
	IsActive   bool      `json:"is_active" example:"true"`
	CreatedAt  time.Time `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// NewBlacklistInfo 从模型创建黑名单信息
func NewBlacklistInfo(blacklist *models.PhoneBlacklist) BlacklistInfo {
	return BlacklistInfo{
		ID:         blacklist.ID,
		UUID:       blacklist.UUID,
		PhoneMD5:   blacklist.PhoneMD5,
		Source:     blacklist.Source,
		Reason:     blacklist.Reason,
		OperatorID: blacklist.OperatorID,
		IsActive:   blacklist.IsActive,
		CreatedAt:  blacklist.CreatedAt,
		UpdatedAt:  blacklist.UpdatedAt,
	}
}

// GetBlacklistResponse 获取黑名单列表响应
type GetBlacklistResponse struct {
	Items      []BlacklistInfo `json:"items"`
	Pagination PaginationInfo  `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	PageSize   int   `json:"page_size" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int64 `json:"total_pages" example:"5"`
}

// CreateApiCredentialRequest 创建API密钥请求
type CreateApiCredentialRequest struct {
	Name        string `json:"name" binding:"required,max=100" example:"测试密钥"`
	Description string `json:"description" example:"用于测试的API密钥"`
	RateLimit   int    `json:"rate_limit" binding:"min=1,max=10000" example:"1000"`
	ExpiresAt   *time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

// ApiCredentialInfo API密钥信息
type ApiCredentialInfo struct {
	ID          uint64     `json:"id" example:"1"`
	UUID        string     `json:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	APIKey      string     `json:"api_key" example:"ak_1234567890abcdef"`
	Name        string     `json:"name" example:"测试密钥"`
	Description string     `json:"description" example:"用于测试的API密钥"`
	RateLimit   int        `json:"rate_limit" example:"1000"`
	Status      string     `json:"status" example:"active"`
	LastUsedAt  *time.Time `json:"last_used_at" example:"2024-01-01T10:00:00Z"`
	ExpiresAt   *time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// NewApiCredentialInfo 从模型创建API密钥信息
func NewApiCredentialInfo(credential *models.BlacklistApiCredential) ApiCredentialInfo {
	return ApiCredentialInfo{
		ID:          credential.ID,
		UUID:        credential.UUID,
		APIKey:      credential.APIKey,
		Name:        credential.Name,
		Description: credential.Description,
		RateLimit:   credential.RateLimit,
		Status:      credential.Status,
		LastUsedAt:  credential.LastUsedAt,
		ExpiresAt:   credential.ExpiresAt,
		CreatedAt:   credential.CreatedAt,
		UpdatedAt:   credential.UpdatedAt,
	}
}

// QueryStatsResponse 查询统计响应
type QueryStatsResponse struct {
	TotalQueries int64   `json:"total_queries" example:"1000"`
	HitCount     int64   `json:"hit_count" example:"150"`
	MissCount    int64   `json:"miss_count" example:"850"`
	HitRate      float64 `json:"hit_rate" example:"15.0"`
	AvgLatency   float64 `json:"avg_latency_ms" example:"5.2"`
}

// MinuteStatsRequest 分钟级统计请求
type MinuteStatsRequest struct {
	Minutes int `form:"minutes,default=5" binding:"min=1,max=60"`
}

// MinuteStatsResponse 分钟级统计响应
type MinuteStatsResponse struct {
	Timestamp    time.Time     `json:"timestamp"`
	TotalQueries int64         `json:"total_queries"`
	HitCount     int64         `json:"hit_count"`
	MissCount    int64         `json:"miss_count"`
	HitRate      float64       `json:"hit_rate"`
	QPS          float64       `json:"qps"`
	AvgLatency   float64       `json:"avg_latency_ms"`
	MinuteData   []MinutePoint `json:"minute_data"`
}

// MinutePoint 每分钟数据点
type MinutePoint struct {
	Minute       string  `json:"minute"`
	TotalQueries int64   `json:"total_queries"`
	HitCount     int64   `json:"hit_count"`
	QPS          float64 `json:"qps"`
	AvgLatency   float64 `json:"avg_latency_ms"`
}

// BatchImportResponse 批量导入响应
type BatchImportResponse struct {
	SuccessCount int      `json:"success_count" example:"100"`
	FailedCount  int      `json:"failed_count" example:"0"`
	FailedItems  []string `json:"failed_items" example:"[]"`
}