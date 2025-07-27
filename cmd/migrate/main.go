// Package main provides database migration utility.
// It handles database schema creation and migration operations.
package main

import (
	"flag"
	"log"

	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/database"
	"github.com/varluffy/shield/pkg/logger"
)

func main() {
	// 解析命令行参数
	var configPath string
	var clean bool
	var addUUIDs bool
	var verifyOnly bool
	var initPermissions bool
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.BoolVar(&clean, "clean", false, "Clean all tables before migration")
	flag.BoolVar(&addUUIDs, "add-uuids", false, "Add UUID columns to existing tables and generate UUIDs")
	flag.BoolVar(&verifyOnly, "verify", false, "Only verify existing UUID data")
	flag.BoolVar(&initPermissions, "init-permissions", false, "Initialize system permissions and roles")
	flag.Parse()

	// 加载配置
	loader := config.NewConfigLoader()
	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建日志器
	logConfig := &logger.LogConfig{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	}

	appLogger, err := logger.NewLoggerWithConfig(logConfig)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// 连接数据库
	db, err := database.NewMySQLConnection(cfg.Database, appLogger.Logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 如果需要清理，先删除所有表
	if clean {
		if err := database.CleanDatabase(db); err != nil {
			log.Fatalf("Failed to clean database: %v", err)
		}
	}

	// 创建迁移实例
	migration := NewMigration(db, appLogger.Logger)

	// 处理UUID相关操作
	if verifyOnly {
		// 只验证UUID数据
		log.Println("🔍 Verifying UUID data...")
		if err := migration.VerifyMigration(); err != nil {
			log.Fatalf("❌ UUID verification failed: %v", err)
		}
		log.Println("✅ UUID verification completed successfully")
		return
	}

	if addUUIDs {
		// 执行UUID迁移
		log.Println("🚀 Starting UUID migration...")
		if err := migration.RunMigration(); err != nil {
			log.Fatalf("❌ UUID migration failed: %v", err)
		}

		// 验证迁移结果
		if err := migration.VerifyMigration(); err != nil {
			log.Fatalf("❌ UUID migration verification failed: %v", err)
		}

		log.Println("🎉 UUID migration completed successfully!")
		return
	}

	if initPermissions {
		// 执行权限系统初始化
		log.Println("🚀 Starting permission system initialization...")
		if err := migration.RunPermissionMigration(); err != nil {
			log.Fatalf("❌ Permission system initialization failed: %v", err)
		}
		log.Println("🎉 Permission system initialization completed successfully!")
		return
	}

	// 执行标准数据库迁移
	if err := database.MigrateDatabase(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
}
