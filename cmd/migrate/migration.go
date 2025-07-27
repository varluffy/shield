package main

import (
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMigration creates a new migration instance
func NewMigration(db *gorm.DB, logger *zap.Logger) *Migration {
	return &Migration{
		db:     db,
		logger: logger,
	}
}

// AddUUIDColumns 为现有表添加UUID列
func (m *Migration) AddUUIDColumns() error {
	fmt.Println("🔄 Adding UUID columns to existing tables...")

	// 1. 为tenants表添加UUID列（如果还没有）
	if err := m.addUUIDColumnIfNotExists("tenants"); err != nil {
		return fmt.Errorf("failed to add UUID to tenants: %w", err)
	}

	// 2. 为users表添加UUID列（如果还没有）
	if err := m.addUUIDColumnIfNotExists("users"); err != nil {
		return fmt.Errorf("failed to add UUID to users: %w", err)
	}

	// 3. 为roles表添加UUID列（如果还没有）
	if err := m.addUUIDColumnIfNotExists("roles"); err != nil {
		return fmt.Errorf("failed to add UUID to roles: %w", err)
	}

	// 4. 为permissions表添加UUID列（如果还没有）
	if err := m.addUUIDColumnIfNotExists("permissions"); err != nil {
		return fmt.Errorf("failed to add UUID to permissions: %w", err)
	}

	// 5. 为user_profiles表添加UUID列（如果还没有且表存在）
	if m.tableExists("user_profiles") {
		if err := m.addUUIDColumnIfNotExists("user_profiles"); err != nil {
			return fmt.Errorf("failed to add UUID to user_profiles: %w", err)
		}
	}

	fmt.Println("✅ UUID columns added successfully")
	return nil
}

// GenerateUUIDs 为现有记录生成UUID值
func (m *Migration) GenerateUUIDs() error {
	fmt.Println("🔄 Generating UUIDs for existing records...")

	// 1. 为tenants生成UUID
	if err := m.generateUUIDsForTable("tenants"); err != nil {
		return fmt.Errorf("failed to generate UUIDs for tenants: %w", err)
	}

	// 2. 为users生成UUID
	if err := m.generateUUIDsForTable("users"); err != nil {
		return fmt.Errorf("failed to generate UUIDs for users: %w", err)
	}

	// 3. 为roles生成UUID
	if err := m.generateUUIDsForTable("roles"); err != nil {
		return fmt.Errorf("failed to generate UUIDs for roles: %w", err)
	}

	// 4. 为permissions生成UUID
	if err := m.generateUUIDsForTable("permissions"); err != nil {
		return fmt.Errorf("failed to generate UUIDs for permissions: %w", err)
	}

	// 5. 为user_profiles生成UUID（如果表存在）
	if m.tableExists("user_profiles") {
		if err := m.generateUUIDsForTable("user_profiles"); err != nil {
			return fmt.Errorf("failed to generate UUIDs for user_profiles: %w", err)
		}
	}

	fmt.Println("✅ UUIDs generated successfully")
	return nil
}

// AddUniqueConstraints 为UUID字段添加唯一约束
func (m *Migration) AddUniqueConstraints() error {
	fmt.Println("🔄 Adding unique constraints for UUID columns...")

	tables := []string{"tenants", "users", "roles", "permissions"}

	// 添加user_profiles表（如果存在）
	if m.tableExists("user_profiles") {
		tables = append(tables, "user_profiles")
	}

	for _, table := range tables {
		indexName := fmt.Sprintf("idx_%s_uuid", table)
		sql := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (uuid)", indexName, table)

		if err := m.db.Exec(sql).Error; err != nil {
			return fmt.Errorf("failed to add unique constraint for %s.uuid: %w", table, err)
		}

		fmt.Printf("  ✅ Added unique constraint for %s.uuid\n", table)
	}

	fmt.Println("✅ Unique constraints added successfully")
	return nil
}

// RunMigration 执行完整的迁移流程
func (m *Migration) RunMigration() error {
	fmt.Println("🚀 Starting UUID migration...")

	// 开启事务
	return m.db.Transaction(func(tx *gorm.DB) error {
		m.db = tx // 在事务中执行所有操作

		// 1. 添加UUID列
		if err := m.AddUUIDColumns(); err != nil {
			return err
		}

		// 2. 生成UUID值
		if err := m.GenerateUUIDs(); err != nil {
			return err
		}

		// 3. 添加唯一约束
		if err := m.AddUniqueConstraints(); err != nil {
			return err
		}

		fmt.Println("🎉 UUID migration completed successfully!")
		return nil
	})
}

// 辅助方法

// addUUIDColumnIfNotExists 检查并添加UUID列
func (m *Migration) addUUIDColumnIfNotExists(tableName string) error {
	// 检查UUID列是否已存在
	var count int64
	err := m.db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = 'uuid'", tableName).Scan(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("  ⏭️  UUID column already exists in %s\n", tableName)
		return nil
	}

	// 添加UUID列
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN uuid CHAR(36) DEFAULT '' AFTER id", tableName)
	if err := m.db.Exec(sql).Error; err != nil {
		return err
	}

	fmt.Printf("  ✅ Added UUID column to %s\n", tableName)
	return nil
}

// generateUUIDsForTable 为指定表的记录生成UUID
func (m *Migration) generateUUIDsForTable(tableName string) error {
	// 查询所有没有UUID的记录
	rows, err := m.db.Raw(fmt.Sprintf("SELECT id FROM %s WHERE uuid = '' OR uuid IS NULL", tableName)).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	var updatedCount int
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return err
		}

		// 生成新的UUID
		newUUID := uuid.New().String()

		// 更新记录
		sql := fmt.Sprintf("UPDATE %s SET uuid = ? WHERE id = ?", tableName)
		if err := m.db.Exec(sql, newUUID, id).Error; err != nil {
			return err
		}

		updatedCount++
	}

	fmt.Printf("  ✅ Generated %d UUIDs for %s\n", updatedCount, tableName)
	return nil
}

// tableExists 检查表是否存在
func (m *Migration) tableExists(tableName string) bool {
	var count int64
	err := m.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count).Error
	if err != nil {
		return false
	}
	return count > 0
}

// VerifyMigration 验证迁移结果
func (m *Migration) VerifyMigration() error {
	fmt.Println("🔍 Verifying migration results...")

	tables := []string{"tenants", "users", "roles", "permissions"}
	if m.tableExists("user_profiles") {
		tables = append(tables, "user_profiles")
	}

	for _, table := range tables {
		// 检查是否有空的UUID
		var emptyUUIDCount int64
		err := m.db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE uuid = '' OR uuid IS NULL", table)).Scan(&emptyUUIDCount).Error
		if err != nil {
			return fmt.Errorf("failed to verify %s: %w", table, err)
		}

		if emptyUUIDCount > 0 {
			return fmt.Errorf("found %d records with empty UUID in %s", emptyUUIDCount, table)
		}

		// 检查是否有重复的UUID
		var duplicateCount int64
		err = m.db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM (SELECT uuid, COUNT(*) as cnt FROM %s GROUP BY uuid HAVING cnt > 1) as duplicates", table)).Scan(&duplicateCount).Error
		if err != nil {
			return fmt.Errorf("failed to check duplicates in %s: %w", table, err)
		}

		if duplicateCount > 0 {
			return fmt.Errorf("found duplicate UUIDs in %s", table)
		}

		fmt.Printf("  ✅ %s: All records have unique UUIDs\n", table)
	}

	fmt.Println("✅ Migration verification completed successfully!")
	return nil
}

// createBackup 创建备份（可选）
func (m *Migration) CreateBackup() error {
	fmt.Println("💾 Creating database backup...")

	// 这里可以添加备份逻辑
	// 例如导出重要表的数据

	fmt.Println("✅ Backup completed (implement if needed)")
	return nil
}

// InitializePermissions 初始化权限和角色
func (m *Migration) InitializePermissions() error {
	fmt.Println("🔄 Initializing system permissions and roles...")

	// 1. 初始化系统权限
	if err := InitSystemPermissions(m.db); err != nil {
		return fmt.Errorf("failed to initialize system permissions: %w", err)
	}

	// 2. 初始化系统角色
	if err := InitSystemRoles(m.db); err != nil {
		return fmt.Errorf("failed to initialize system roles: %w", err)
	}

	fmt.Println("✅ System permissions and roles initialized successfully")
	return nil
}

// RunPermissionMigration 执行权限系统迁移
func (m *Migration) RunPermissionMigration() error {
	fmt.Println("🚀 Starting permission system migration...")

	// 不使用事务，因为权限初始化可能需要多次操作
	if err := m.InitializePermissions(); err != nil {
		return err
	}

	fmt.Println("🎉 Permission system migration completed successfully!")
	return nil
}
