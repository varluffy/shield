package models

import (
	"time"

	"gorm.io/gorm"
)

// PhoneBlacklist 手机号黑名单模型
type PhoneBlacklist struct {
	TenantModel
	PhoneMD5   string `gorm:"type:char(32);not null;uniqueIndex:uk_tenant_phone_md5" json:"phone_md5"`
	Source     string `gorm:"type:varchar(50);not null" json:"source"` // manual, import, api
	Reason     string `gorm:"type:varchar(200)" json:"reason"`         // 加入黑名单原因
	OperatorID uint64 `gorm:"index" json:"operator_id"`                // 操作人ID
	IsActive   bool   `gorm:"default:true" json:"is_active"`           // 是否有效
}

func (PhoneBlacklist) TableName() string {
	return "phone_blacklists"
}

// BeforeCreate 创建前钩子
func (pb *PhoneBlacklist) BeforeCreate(tx *gorm.DB) error {
	if pb.UUID == "" {
		pb.UUID = GenerateUUID()
	}
	// 从上下文获取租户ID
	if pb.TenantID == 0 {
		pb.TenantID = GetTenantIDFromContext(tx)
	}
	return nil
}

// BlacklistApiCredential 黑名单API密钥模型
type BlacklistApiCredential struct {
	TenantModel
	APIKey      string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"api_key"`
	APISecret   string     `gorm:"type:varchar(128);not null" json:"api_secret"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`          // 密钥名称
	Description string     `gorm:"type:text" json:"description"`                    // 描述
	RateLimit   int        `gorm:"default:1000" json:"rate_limit"`                  // 每秒请求限制
	IPWhitelist string     `gorm:"type:text" json:"ip_whitelist"`                   // IP白名单，逗号分隔，支持CIDR
	Status      string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, suspended
	LastUsedAt  *time.Time `json:"last_used_at"`                                    // 最后使用时间
	ExpiresAt   *time.Time `json:"expires_at"`                                      // 过期时间
}

func (BlacklistApiCredential) TableName() string {
	return "blacklist_api_credentials"
}

// BeforeCreate 创建前钩子
func (bac *BlacklistApiCredential) BeforeCreate(tx *gorm.DB) error {
	if bac.UUID == "" {
		bac.UUID = GenerateUUID()
	}
	// 从上下文获取租户ID
	if bac.TenantID == 0 {
		bac.TenantID = GetTenantIDFromContext(tx)
	}
	return nil
}

// BlacklistQueryLog 黑名单查询日志模型（用于统计分析）
type BlacklistQueryLog struct {
	BaseModelWithoutUUID
	TenantID     uint64 `gorm:"not null;index" json:"tenant_id"`
	APIKey       string `gorm:"type:varchar(64);not null;index" json:"api_key"`
	PhoneMD5     string `gorm:"type:char(32);not null" json:"phone_md5"`
	IsHit        bool   `gorm:"default:false" json:"is_hit"`              // 是否命中黑名单
	ResponseTime int    `gorm:"not null" json:"response_time"`            // 响应时间(毫秒)
	ClientIP     string `gorm:"type:varchar(45)" json:"client_ip"`        // 客户端IP
	UserAgent    string `gorm:"type:text" json:"user_agent"`              // 用户代理
	RequestID    string `gorm:"type:varchar(64);index" json:"request_id"` // 请求ID
}

func (BlacklistQueryLog) TableName() string {
	return "blacklist_query_logs"
}