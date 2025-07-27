// Package database provides database connection and configuration.
// It handles MySQL database initialization and connection management.
package database

import (
	"fmt"
	"time"

	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/models"
	pkgLogger "github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

// NewMySQLConnection 创建MySQL连接，支持OpenTelemetry追踪和自定义Zap日志
func NewMySQLConnection(cfg config.DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	// 创建自定义GORM日志器
	var gormLogger *pkgLogger.GormLogger
	if zapLogger != nil {
		// 解析日志级别
		logLevel := pkgLogger.ParseGormLogLevel(cfg.LogLevel)

		// 创建GORM logger配置
		gormLoggerConfig := pkgLogger.GormLoggerConfig{
			LogLevel:                  logLevel,
			SlowThreshold:             cfg.SlowQueryThreshold,
			SkipCallerLookup:          false,
			IgnoreRecordNotFoundError: true,
		}

		gormLogger = pkgLogger.NewGormLoggerWithConfig(zapLogger, gormLoggerConfig)
		zapLogger.Info("Created custom GORM logger with Zap and TraceID support",
			zap.String("log_level", cfg.LogLevel),
			zap.Duration("slow_query_threshold", cfg.SlowQueryThreshold),
		)
	}

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 添加OpenTelemetry追踪插件
	if zapLogger != nil {
		if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
			zapLogger.Warn("Failed to enable database tracing", zap.Error(err))
		} else {
			zapLogger.Info("Database tracing enabled successfully")
		}
	}

	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Permission{},
		&models.Role{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.RefreshToken{},
		&models.LoginAttempt{},
		&models.UserProfile{},
	)
}

// MigrateDatabase 迁移数据库
func MigrateDatabase(db *gorm.DB) error {
	fmt.Println("Starting database migration...")

	// 执行自动迁移
	if err := AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化系统数据
	if err := SeedInitialData(db); err != nil {
		return fmt.Errorf("failed to seed initial data: %w", err)
	}

	fmt.Println("Database migration completed successfully")
	return nil
}

// SeedInitialData 初始化系统数据
func SeedInitialData(db *gorm.DB) error {
	fmt.Println("Seeding initial data...")

	// 创建默认租户
	defaultTenant := &models.Tenant{
		Name:     "默认租户",
		Domain:   "default.ultrafit.com",
		Status:   "active",
		Plan:     "basic",
		Settings: "{}",
	}

	// 检查是否已存在默认租户
	var existingTenant models.Tenant
	if err := db.Where("domain = ?", defaultTenant.Domain).First(&existingTenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(defaultTenant).Error; err != nil {
				return fmt.Errorf("failed to create default tenant: %w", err)
			}
			fmt.Printf("Created default tenant: %d\n", defaultTenant.ID)
		} else {
			return fmt.Errorf("failed to check default tenant: %w", err)
		}
	} else {
		defaultTenant = &existingTenant
		fmt.Printf("Default tenant already exists: %d\n", defaultTenant.ID)
	}

	// 创建系统权限
	if err := createSystemPermissions(db); err != nil {
		return fmt.Errorf("failed to create system permissions: %w", err)
	}

	// 创建系统角色
	if err := createSystemRoles(db, defaultTenant.ID); err != nil {
		return fmt.Errorf("failed to create system roles: %w", err)
	}

	fmt.Println("Initial data seeded successfully")
	return nil
}

// createSystemPermissions 创建系统权限
// createSystemPermissions 已弃用，现在使用 cmd/migrate/permissions.go 中的权限初始化逻辑
func createSystemPermissions(db *gorm.DB) error {
	// 旧的权限初始化逻辑已移至 cmd/migrate/permissions.go
	// 这里只是一个占位符，实际权限初始化通过 migration 工具完成
	return nil
}

// createSystemRoles 已弃用，现在使用 cmd/migrate/permissions.go 中的角色初始化逻辑
func createSystemRoles(db *gorm.DB, tenantID uint64) error {
	// 旧的角色初始化逻辑已移至 cmd/migrate/permissions.go
	// 这里只是一个占位符，实际角色初始化通过 migration 工具完成
	return nil
}

// assignRolePermissions 已弃用，现在使用 cmd/migrate/permissions.go 中的权限分配逻辑
func assignRolePermissions(db *gorm.DB, roleID uint64, roleCode string) error {
	// 旧的权限分配逻辑已移至 cmd/migrate/permissions.go
	// 这里只是一个占位符，实际权限分配通过 migration 工具完成
	return nil
}

// CleanDatabase 清理数据库中的所有表
func CleanDatabase(db *gorm.DB) error {
	fmt.Println("Cleaning database...")

	// 禁用外键检查
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// 删除表的顺序很重要，先删除有外键约束的表
	tables := []string{
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
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			fmt.Printf("Warning: Failed to drop table %s: %v\n", table, err)
		} else {
			fmt.Printf("Dropped table: %s\n", table)
		}
	}

	// 重新启用外键检查
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		return fmt.Errorf("failed to enable foreign key checks: %w", err)
	}

	fmt.Println("Database cleaned successfully")
	return nil
}
