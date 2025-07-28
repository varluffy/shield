// Package test provides test utilities and helpers for the application.
package test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mojocn/base64Captcha/store"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/database"
	"github.com/varluffy/shield/internal/handlers"
	"github.com/varluffy/shield/internal/middleware"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/auth"
	"github.com/varluffy/shield/pkg/captcha"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/redis"
	"github.com/varluffy/shield/pkg/response"
	"github.com/varluffy/shield/pkg/transaction"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TestComponents 测试组件集合
type TestComponents struct {
	// Repositories
	UserRepo            repositories.UserRepository
	RoleRepo            repositories.RoleRepository
	PermissionRepo      repositories.PermissionRepository
	TenantRepo          repositories.TenantRepository
	PermissionAuditRepo repositories.PermissionAuditRepository

	// Services
	UserService            services.UserService
	PermissionService      services.PermissionService
	RoleService            services.RoleService
	FieldPermissionService services.FieldPermissionService
	PermissionCacheService services.PermissionCacheService
	PermissionAuditService services.PermissionAuditService

	// Handlers
	UserHandler            *handlers.UserHandler
	PermissionHandler      *handlers.PermissionHandler
	RoleHandler            *handlers.RoleHandler
	FieldPermissionHandler *handlers.FieldPermissionHandler
	CaptchaHandler         *handlers.CaptchaHandler

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware

	// Infrastructure
	JWTService     auth.JWTService
	CaptchaService captcha.CaptchaService
	TxManager      transaction.TransactionManager
}

// NewTestConfig 创建测试配置
func NewTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "shield-test",
			Version:     "1.0.0",
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "123456",
			Name:     "shield_test",
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
			CORS: config.CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
				AllowHeaders: []string{"*"},
			},
		},
		Auth: &config.AuthConfig{
			JWT: config.JWTConfig{
				Secret:         "test-secret-key",
				Issuer:         "shield-test",
				ExpiresIn:      time.Hour,
				RefreshExpires: time.Hour * 24,
			},
		},
		Redis: &config.RedisConfig{
			Addrs:    []string{"localhost:6379"},
			Password: "",
			DB:       1, // 使用DB 1用于测试
		},
		Jaeger: &config.JaegerConfig{
			Enabled:    false, // 测试时禁用Jaeger
			OTLPURL:    "http://localhost:4318/v1/traces",
			SampleRate: 1.0,
		},
	}
}

// NewTestLogger 创建测试日志器
func NewTestLogger() (*logger.Logger, error) {
	return logger.NewLogger("test")
}

// NewTestComponents 创建完整的测试组件
func NewTestComponents(db *gorm.DB, testLogger *logger.Logger) *TestComponents {
	// 创建事务管理器
	txManager := transaction.NewTransactionManager(db, testLogger.Logger)

	// 创建JWT服务
	jwtService := auth.NewJWTService(
		"test-secret-key",
		"shield-test",
		time.Hour,
		time.Hour*24,
	)

	// 创建Captcha服务
	captchaStore := store.NewMemoryStore(1000, 240*time.Second)
	captchaConfig := &captcha.CaptchaConfig{
		Type:       "digit",
		Width:      240,
		Height:     80,
		Length:     4,
		NoiseCount: 0,
	}
	captchaService := captcha.NewCaptchaService(captchaStore, captchaConfig, testLogger.Logger)

	// 创建Redis客户端（可选，如果Redis不可用则跳过缓存）
	redisCache := redis.NewClient(&redis.Config{
		Addrs:    []string{"localhost:6379"},
		Password: "",
		DB:       1,
	}, testLogger.Logger)

	// 创建Repositories
	userRepo := repositories.NewUserRepository(db, txManager, testLogger)
	roleRepo := repositories.NewRoleRepository(db, txManager, testLogger)
	permissionRepo := repositories.NewPermissionRepository(db, txManager, testLogger)
	tenantRepo := repositories.NewTenantRepository(db, txManager, testLogger)
	permissionAuditRepo := repositories.NewPermissionAuditRepository(db, txManager, testLogger)

	// 创建Services
	var permissionCacheService services.PermissionCacheService
	if redisCache != nil {
		permissionCacheService = services.NewPermissionCacheService(redisCache, testLogger)
	}

	permissionService := services.NewPermissionService(
		userRepo, roleRepo, permissionRepo, tenantRepo, permissionCacheService, testLogger,
	)

	userService := services.NewUserService(userRepo, testLogger, txManager, jwtService, captchaService)
	roleService := services.NewRoleService(roleRepo, permissionRepo, testLogger)
	fieldPermissionService := services.NewFieldPermissionService(testLogger)
	permissionAuditService := services.NewPermissionAuditService(permissionAuditRepo, testLogger)

	// 创建ResponseWriter
	responseWriter := response.NewResponseWriter(testLogger)

	// 创建Handlers
	userHandler := handlers.NewUserHandler(userService, permissionService, testLogger)
	permissionHandler := handlers.NewPermissionHandler(permissionService, testLogger)
	roleHandler := handlers.NewRoleHandler(roleService, testLogger)
	fieldPermissionHandler := handlers.NewFieldPermissionHandler(fieldPermissionService, testLogger)
	captchaHandler := handlers.NewCaptchaHandler(captchaService, responseWriter, testLogger.Logger)

	// 创建Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, permissionService, testLogger)

	return &TestComponents{
		// Repositories
		UserRepo:            userRepo,
		RoleRepo:            roleRepo,
		PermissionRepo:      permissionRepo,
		TenantRepo:          tenantRepo,
		PermissionAuditRepo: permissionAuditRepo,

		// Services
		UserService:            userService,
		PermissionService:      permissionService,
		RoleService:            roleService,
		FieldPermissionService: fieldPermissionService,
		PermissionCacheService: permissionCacheService,
		PermissionAuditService: permissionAuditService,

		// Handlers
		UserHandler:            userHandler,
		PermissionHandler:      permissionHandler,
		RoleHandler:            roleHandler,
		FieldPermissionHandler: fieldPermissionHandler,
		CaptchaHandler:         captchaHandler,

		// Middleware
		AuthMiddleware: authMiddleware,

		// Infrastructure
		JWTService:     jwtService,
		CaptchaService: captchaService,
		TxManager:      txManager,
	}
}

// SetupTestDB 设置测试数据库
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	cfg := NewTestConfig()
	testLogger, err := NewTestLogger()
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}

	db, err := database.NewMySQLConnection(cfg.Database, testLogger.Logger)
	if err != nil {
		t.Skipf("Skip test due to database connection error: %v", err)
		return nil, nil
	}

	// 自动迁移
	if err := database.AutoMigrate(db); err != nil {
		t.Skipf("Skip test due to migration error: %v", err)
		return nil, nil
	}

	// 清理函数
	cleanup := func() {
		// 清理测试数据
		CleanupTestData(db)

		// 关闭数据库连接
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

// CreateTestTenant 创建测试租户
func CreateTestTenant(db *gorm.DB) *models.Tenant {
	tenant := &models.Tenant{
		Name:       "Test Tenant",
		Domain:     "test.example.com",
		Status:     "active",
		Plan:       "basic",
		MaxUsers:   100,
		MaxStorage: 1073741824, // 1GB
	}

	if err := db.Create(tenant).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test tenant: %v", err))
	}

	return tenant
}

// CreateTestUser 创建测试用户
func CreateTestUser(db *gorm.DB, tenantID uint64, email string) *models.User {
	user := &models.User{
		TenantModel: models.TenantModel{TenantID: tenantID},
		Email:       email,
		Password:    "$2a$10$example.hashed.password", // 模拟加密密码
		Name:        "Test User",
		Status:      "active",
	}

	if err := db.Create(user).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test user: %v", err))
	}

	return user
}

// CreateSystemAdmin 创建系统管理员用户
func CreateSystemAdmin(db *gorm.DB) *models.User {
	return CreateTestUser(db, 0, "admin@system.test") // tenant_id = 0 表示系统租户
}

// CreateTestRole 创建测试角色
func CreateTestRole(db *gorm.DB, tenantID uint64, code, name string) *models.Role {
	role := &models.Role{
		TenantModel: models.TenantModel{TenantID: tenantID},
		Code:        code,
		Name:        name,
		Type:        "custom",
		IsActive:    true,
	}

	if err := db.Create(role).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test role: %v", err))
	}

	return role
}

// CreateTestPermission 创建测试权限
func CreateTestPermission(db *gorm.DB, code, name, scope, permType string) *models.Permission {
	permission := &models.Permission{
		Code:        code,
		Name:        name,
		Description: fmt.Sprintf("Test permission: %s", name),
		Type:        permType,
		Scope:       scope,
		IsBuiltin:   false,
		IsActive:    true,
		Module:      "test",
	}

	if err := db.Create(permission).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test permission: %v", err))
	}

	return permission
}

// CleanupTestData 清理测试数据
func CleanupTestData(db *gorm.DB) {
	// 按照外键关系的反向顺序删除
	tables := []string{
		"permission_audit_logs",
		"role_field_permissions",
		"field_permissions",
		"user_roles",
		"role_permissions",
		"refresh_tokens",
		"login_attempts",
		"user_profiles",
		"users",
		"roles",
		"permissions",
		"tenants",
	}

	for _, table := range tables {
		db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id > 0", table))
	}
}

// SeedTestData 种子测试数据
func SeedTestData(db *gorm.DB) {
	// 创建系统租户用户
	systemAdmin := CreateSystemAdmin(db)

	// 创建测试租户
	tenant := CreateTestTenant(db)

	// 创建租户用户
	tenantUser := CreateTestUser(db, tenant.ID, "user@tenant.test")

	// 创建测试权限
	systemPerm := CreateTestPermission(db, "test_system_perm", "测试系统权限", "system", "api")
	tenantPerm := CreateTestPermission(db, "test_tenant_perm", "测试租户权限", "tenant", "api")

	// 创建测试角色
	systemRole := CreateTestRole(db, 0, "system_admin", "系统管理员")
	tenantRole := CreateTestRole(db, tenant.ID, "tenant_admin", "租户管理员")

	// 分配权限给角色
	db.Create(&models.RolePermission{
		RoleID:       systemRole.ID,
		PermissionID: systemPerm.ID,
		GrantedBy:    systemAdmin.ID,
	})

	db.Create(&models.RolePermission{
		RoleID:       tenantRole.ID,
		PermissionID: tenantPerm.ID,
		GrantedBy:    systemAdmin.ID,
	})

	// 分配角色给用户
	db.Create(&models.UserRole{
		UserID:    systemAdmin.ID,
		RoleID:    systemRole.ID,
		TenantID:  0,
		GrantedBy: systemAdmin.ID,
		IsActive:  true,
	})

	db.Create(&models.UserRole{
		UserID:    tenantUser.ID,
		RoleID:    tenantRole.ID,
		TenantID:  tenant.ID,
		GrantedBy: systemAdmin.ID,
		IsActive:  true,
	})
}

// =============================================================================
// 认证辅助方法
// =============================================================================

// StandardTestUser 标准测试用户配置
type StandardTestUser struct {
	Email     string
	Password  string
	Name      string
	TenantID  uint64
	RoleCode  string
	IsActive  bool
}

// GetStandardTestUsers 获取标准测试用户配置
func GetStandardTestUsers() []StandardTestUser {
	return []StandardTestUser{
		{
			Email:    "admin@system.test",
			Password: "admin123",
			Name:     "系统管理员",
			TenantID: 0, // 系统租户
			RoleCode: "system_admin",
			IsActive: true,
		},
		{
			Email:    "admin@tenant.test",
			Password: "admin123",
			Name:     "租户管理员",
			TenantID: 1, // 默认租户
			RoleCode: "tenant_admin",
			IsActive: true,
		},
		{
			Email:    "user@tenant.test",
			Password: "user123",
			Name:     "普通用户",
			TenantID: 1, // 默认租户
			RoleCode: "user",
			IsActive: true,
		},
		{
			Email:    "test@example.com",
			Password: "test123",
			Name:     "测试用户",
			TenantID: 1, // 默认租户
			RoleCode: "user",
			IsActive: true,
		},
	}
}

// HashTestPassword 生成测试密码的BCrypt哈希
func HashTestPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("Failed to hash test password: %v", err))
	}
	return string(hashedPassword)
}

// GenerateTestJWT 为测试用户生成JWT Token
func GenerateTestJWT(components *TestComponents, userID, tenantID string) (string, error) {
	// 生成JWT Token - 需要用户邮箱，先构造一个测试邮箱
	testEmail := fmt.Sprintf("test-user-%s@example.com", userID)
	
	token, err := components.JWTService.GenerateAccessToken(userID, testEmail, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}
	
	return token, nil
}

// CreateAuthHeader 创建认证请求头
func CreateAuthHeader(token string) http.Header {
	header := make(http.Header)
	header.Set("Authorization", "Bearer "+token)
	header.Set("Content-Type", "application/json")
	return header
}

// LoginAsTestUser 使用标准测试用户快速登录
func LoginAsTestUser(components *TestComponents, email string) (string, error) {
	// 查找标准测试用户
	testUsers := GetStandardTestUsers()
	var targetUser *StandardTestUser
	for _, user := range testUsers {
		if user.Email == email {
			targetUser = &user
			break
		}
	}
	
	if targetUser == nil {
		return "", fmt.Errorf("test user with email %s not found", email)
	}
	
	// 生成JWT Token
	userUUID := fmt.Sprintf("test-user-%d", targetUser.TenantID)
	tenantIDStr := strconv.FormatUint(targetUser.TenantID, 10)
	
	return GenerateTestJWT(components, userUUID, tenantIDStr)
}

// CreateStandardTestUser 创建标准测试用户实例
func CreateStandardTestUser(db *gorm.DB, testUser StandardTestUser) *models.User {
	// 检查用户是否已存在
	var existingUser models.User
	err := db.Where("tenant_id = ? AND email = ?", testUser.TenantID, testUser.Email).First(&existingUser).Error
	if err == nil {
		return &existingUser // 用户已存在，直接返回
	}
	
	// 创建新用户
	user := &models.User{
		TenantModel: models.TenantModel{TenantID: testUser.TenantID},
		Email:       testUser.Email,
		Password:    HashTestPassword(testUser.Password),
		Name:        testUser.Name,
		Status:      "active",
	}
	
	if testUser.IsActive {
		user.Status = "active"
	} else {
		user.Status = "inactive"
	}
	
	if err := db.Create(user).Error; err != nil {
		panic(fmt.Sprintf("Failed to create standard test user %s: %v", testUser.Email, err))
	}
	
	return user
}

// SetupStandardTestUsers 设置所有标准测试用户
func SetupStandardTestUsers(db *gorm.DB) map[string]*models.User {
	testUsers := GetStandardTestUsers()
	userMap := make(map[string]*models.User)
	
	for _, testUser := range testUsers {
		user := CreateStandardTestUser(db, testUser)
		userMap[testUser.Email] = user
	}
	
	return userMap
}
