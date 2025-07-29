package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/varluffy/shield/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// MigrationRunner 迁移运行器
type MigrationRunner struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMigrationRunner 创建迁移运行器
func NewMigrationRunner(db *gorm.DB, logger *zap.Logger) *MigrationRunner {
	return &MigrationRunner{
		db:     db,
		logger: logger,
	}
}

// RunMigrations 运行迁移
func (r *MigrationRunner) RunMigrations() error {
	// 初始化迁移管理器
	manager := database.NewMigrationManager(r.db, r.logger)
	if err := manager.InitMigrationTable(); err != nil {
		return err
	}

	// 加载迁移文件
	loader := database.NewMigrationLoader("migrations", r.logger)
	migrations, err := loader.LoadMigrations(migrationFS)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// 验证迁移文件
	if err := loader.ValidateMigrationFiles(migrations); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// 获取待执行的迁移
	pending, err := manager.GetPendingMigrations(migrations)
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pending) == 0 {
		r.logger.Info("No pending migrations found")
		return nil
	}

	r.logger.Info("Found pending migrations", zap.Int("count", len(pending)))

	// 计算下一个批次号
	executed, err := manager.GetExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	nextBatch := 1
	if len(executed) > 0 {
		// 找到最大批次号
		for _, migration := range executed {
			if migration.Batch >= nextBatch {
				nextBatch = migration.Batch + 1
			}
		}
	}

	// 执行迁移
	for _, migration := range pending {
		if err := manager.ExecuteMigration(migration, nextBatch); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}
	}

	r.logger.Info("All migrations executed successfully", 
		zap.Int("executed", len(pending)),
		zap.Int("batch", nextBatch))

	return nil
}

// RollbackLastBatch 回滚最后一个批次
func (r *MigrationRunner) RollbackLastBatch() error {
	manager := database.NewMigrationManager(r.db, r.logger)

	// 获取已执行的迁移
	executed, err := manager.GetExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	if len(executed) == 0 {
		r.logger.Info("No migrations to rollback")
		return nil
	}

	// 找到最后一个批次
	lastBatch := 0
	for _, migration := range executed {
		if migration.Batch > lastBatch {
			lastBatch = migration.Batch
		}
	}

	// 找到最后一个批次的所有迁移（按版本倒序）
	var toRollback []database.Migration
	for i := len(executed) - 1; i >= 0; i-- {
		if executed[i].Batch == lastBatch {
			toRollback = append(toRollback, executed[i])
		}
	}

	r.logger.Info("Rolling back migrations", 
		zap.Int("count", len(toRollback)),
		zap.Int("batch", lastBatch))

	// 执行回滚
	for _, migration := range toRollback {
		if err := manager.RollbackMigration(migration.Version); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
		}
	}

	r.logger.Info("Rollback completed successfully", 
		zap.Int("rolled_back", len(toRollback)),
		zap.Int("batch", lastBatch))

	return nil
}

// ShowMigrationStatus 显示迁移状态
func (r *MigrationRunner) ShowMigrationStatus() error {
	manager := database.NewMigrationManager(r.db, r.logger)

	// 加载迁移文件
	loader := database.NewMigrationLoader("migrations", r.logger)
	migrations, err := loader.LoadMigrations(migrationFS)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// 获取状态
	status, err := manager.GetMigrationStatus(migrations)
	if err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	// 打印状态
	fmt.Printf("Migration Status:\n")
	fmt.Printf("  Total migrations: %d\n", status.Total)
	fmt.Printf("  Executed: %d\n", status.Executed)
	fmt.Printf("  Pending: %d\n", status.Pending)
	fmt.Printf("\n")

	if len(status.Migrations) > 0 {
		fmt.Printf("Executed migrations:\n")
		for _, migration := range status.Migrations {
			fmt.Printf("  [%d] %s - %s (%s)\n", 
				migration.Batch,
				migration.Version, 
				migration.Name,
				migration.ExecutedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("\n")
	}

	if len(status.PendingList) > 0 {
		fmt.Printf("Pending migrations:\n")
		for _, migration := range status.PendingList {
			fmt.Printf("  %s - %s\n", migration.Version, migration.Name)
		}
	}

	return nil
}

// CreateMigrationFile 创建新的迁移文件
func (r *MigrationRunner) CreateMigrationFile(name string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}

	// 生成版本号
	version := database.GenerateMigrationVersion()
	fileName := fmt.Sprintf("%s_%s.sql", version, name)
	
	// 确保迁移目录存在
	migrationDir := "cmd/migrate/migrations"
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create migration directory: %w", err)
	}

	filePath := filepath.Join(migrationDir, fileName)

	// 创建模板内容
	template := fmt.Sprintf(`-- Description: %s
-- Created: %s

-- +migrate Up
-- Add your UP migration SQL here


-- +migrate Down
-- Add your DOWN migration SQL here (for rollback)

`, name, version)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("Created migration file: %s\n", filePath)
	r.logger.Info("Migration file created", 
		zap.String("file", filePath),
		zap.String("version", version),
		zap.String("name", name))

	return nil
}