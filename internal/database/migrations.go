package database

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migration 数据库迁移记录
type Migration struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Version     string    `gorm:"type:varchar(20);not null;uniqueIndex" json:"version"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Batch       int       `gorm:"not null;index" json:"batch"`
	ExecutedAt  time.Time `gorm:"not null" json:"executed_at"`
	RollbackSQL string    `gorm:"type:text" json:"rollback_sql,omitempty"`
}

func (Migration) TableName() string {
	return "schema_migrations"
}

// MigrationFile 迁移文件结构
type MigrationFile struct {
	Version     string
	Name        string
	UpSQL       string
	DownSQL     string
	Description string
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(db *gorm.DB, logger *zap.Logger) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger,
	}
}

// InitMigrationTable 初始化迁移表
func (m *MigrationManager) InitMigrationTable() error {
	if err := m.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}
	
	m.logger.Info("Migration table initialized successfully")
	return nil
}

// GetExecutedMigrations 获取已执行的迁移
func (m *MigrationManager) GetExecutedMigrations() ([]Migration, error) {
	var migrations []Migration
	if err := m.db.Order("version ASC").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}
	return migrations, nil
}

// GetPendingMigrations 获取待执行的迁移
func (m *MigrationManager) GetPendingMigrations(availableMigrations []MigrationFile) ([]MigrationFile, error) {
	executed, err := m.GetExecutedMigrations()
	if err != nil {
		return nil, err
	}
	
	// 创建已执行版本的映射
	executedVersions := make(map[string]bool)
	for _, migration := range executed {
		executedVersions[migration.Version] = true
	}
	
	// 找出未执行的迁移
	var pending []MigrationFile
	for _, migration := range availableMigrations {
		if !executedVersions[migration.Version] {
			pending = append(pending, migration)
		}
	}
	
	// 按版本号排序
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})
	
	return pending, nil
}

// ExecuteMigration 执行单个迁移
func (m *MigrationManager) ExecuteMigration(migration MigrationFile, batch int) error {
	m.logger.Info("Executing migration", 
		zap.String("version", migration.Version), 
		zap.String("name", migration.Name),
		zap.Int("batch", batch))
	
	// 开始事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			m.logger.Error("Migration panicked, rolled back", 
				zap.String("version", migration.Version),
				zap.Any("error", r))
		}
	}()
	
	// 执行迁移SQL
	if migration.UpSQL != "" {
		statements := splitSQL(migration.UpSQL)
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
			}
		}
	}
	
	// 记录迁移历史
	migrationRecord := Migration{
		Version:     migration.Version,
		Name:        migration.Name,
		Batch:       batch,
		ExecutedAt:  time.Now(),
		RollbackSQL: migration.DownSQL,
	}
	
	if err := tx.Create(&migrationRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}
	
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}
	
	m.logger.Info("Migration executed successfully", 
		zap.String("version", migration.Version),
		zap.String("name", migration.Name))
	
	return nil
}

// RollbackMigration 回滚迁移
func (m *MigrationManager) RollbackMigration(version string) error {
	m.logger.Info("Rolling back migration", zap.String("version", version))
	
	// 获取迁移记录
	var migration Migration
	if err := m.db.Where("version = ?", version).First(&migration).Error; err != nil {
		return fmt.Errorf("migration %s not found: %w", version, err)
	}
	
	// 开始事务
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			m.logger.Error("Rollback panicked", 
				zap.String("version", version),
				zap.Any("error", r))
		}
	}()
	
	// 执行回滚SQL
	if migration.RollbackSQL != "" {
		statements := splitSQL(migration.RollbackSQL)
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to rollback migration %s: %w", version, err)
			}
		}
	}
	
	// 删除迁移记录
	if err := tx.Delete(&migration).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete migration record: %w", err)
	}
	
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}
	
	m.logger.Info("Migration rolled back successfully", zap.String("version", version))
	return nil
}

// GetMigrationStatus 获取迁移状态
func (m *MigrationManager) GetMigrationStatus(availableMigrations []MigrationFile) (*MigrationStatus, error) {
	executed, err := m.GetExecutedMigrations()
	if err != nil {
		return nil, err
	}
	
	pending, err := m.GetPendingMigrations(availableMigrations)
	if err != nil {
		return nil, err
	}
	
	status := &MigrationStatus{
		Total:       len(availableMigrations),
		Executed:    len(executed),
		Pending:     len(pending),
		Migrations:  executed,
		PendingList: pending,
	}
	
	return status, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Total       int             `json:"total"`
	Executed    int             `json:"executed"`
	Pending     int             `json:"pending"`
	Migrations  []Migration     `json:"migrations"`
	PendingList []MigrationFile `json:"pending_list"`
}

// splitSQL 分割SQL语句
func splitSQL(sql string) []string {
	// 简单的SQL分割，以分号为分隔符
	// 注意：这是一个简化的实现，实际应用中可能需要更复杂的解析
	statements := strings.Split(sql, ";")
	var result []string
	
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			result = append(result, stmt)
		}
	}
	
	return result
}

// GenerateMigrationVersion 生成迁移版本号
func GenerateMigrationVersion() string {
	return time.Now().Format("20060102_150405")
}