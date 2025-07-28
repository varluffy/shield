package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	TenantModel
	Email               string     `gorm:"type:varchar(255);not null;uniqueIndex:uk_tenant_email" json:"email"`
	Password            string     `gorm:"type:varchar(255);not null" json:"-"`
	Name                string     `gorm:"type:varchar(100)" json:"name"`
	Avatar              string     `gorm:"type:varchar(500)" json:"avatar"`
	Phone               string     `gorm:"type:varchar(20)" json:"phone"`
	Status              string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, locked
	EmailVerifiedAt     *time.Time `json:"email_verified_at"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	LoginCount          int        `gorm:"default:0" json:"login_count"`
	FailedLoginAttempts int        `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until"`
	Timezone            string     `gorm:"type:varchar(50);default:'Asia/Shanghai'" json:"timezone"`
	Language            string     `gorm:"type:varchar(10);default:'zh'" json:"language"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = GenerateUUID()
	}
	// 从上下文获取租户ID
	if u.TenantID == 0 {
		u.TenantID = GetTenantIDFromContext(tx)
	}
	return nil
}

// UserProfile 用户资料模型（保留UUID，可能对外暴露）
type UserProfile struct {
	BaseModel
	UserID    uint64     `gorm:"not null;uniqueIndex" json:"user_id"`
	FirstName string     `gorm:"type:varchar(50)" json:"first_name"`
	LastName  string     `gorm:"type:varchar(50)" json:"last_name"`
	Bio       string     `gorm:"type:text" json:"bio"`
	Birthday  *time.Time `json:"birthday"`
	Gender    string     `gorm:"type:varchar(10)" json:"gender"`
	Address   string     `gorm:"type:text" json:"address"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// BeforeCreate 创建前钩子
func (up *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if up.UUID == "" {
		up.UUID = GenerateUUID()
	}
	return nil
}
