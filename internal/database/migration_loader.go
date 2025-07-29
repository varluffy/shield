package database

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// MigrationLoader 迁移文件加载器
type MigrationLoader struct {
	migrationDir string
	logger       *zap.Logger
}

// NewMigrationLoader 创建迁移文件加载器
func NewMigrationLoader(migrationDir string, logger *zap.Logger) *MigrationLoader {
	return &MigrationLoader{
		migrationDir: migrationDir,
		logger:       logger,
	}
}

// LoadMigrations 加载所有迁移文件
func (l *MigrationLoader) LoadMigrations(fsys fs.FS) ([]MigrationFile, error) {
	var migrations []MigrationFile
	
	// 遍历迁移目录
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// 只处理.sql文件
		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}
		
		// 解析文件名
		migration, err := l.parseMigrationFile(fsys, path)
		if err != nil {
			l.logger.Warn("Failed to parse migration file", 
				zap.String("file", path), 
				zap.Error(err))
			return nil // 跳过无效文件，不终止整个过程
		}
		
		migrations = append(migrations, *migration)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk migration directory: %w", err)
	}
	
	// 按版本号排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	
	l.logger.Info("Loaded migrations", zap.Int("count", len(migrations)))
	return migrations, nil
}

// parseMigrationFile 解析迁移文件
func (l *MigrationLoader) parseMigrationFile(fsys fs.FS, path string) (*MigrationFile, error) {
	// 解析文件名格式: YYYYMMDD_HHMMSS_name.sql
	fileName := filepath.Base(path)
	
	// 匹配版本号和名称
	re := regexp.MustCompile(`^(\d{8}_\d{6})_(.+)\.sql$`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid migration file name format: %s (expected: YYYYMMDD_HHMMSS_name.sql)", fileName)
	}
	
	version := matches[1]
	name := matches[2]
	
	// 读取文件内容
	content, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file %s: %w", path, err)
	}
	
	// 解析UP和DOWN SQL
	upSQL, downSQL, description, err := l.parseSQL(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL content in %s: %w", path, err)
	}
	
	migration := &MigrationFile{
		Version:     version,
		Name:        name,
		UpSQL:       upSQL,
		DownSQL:     downSQL,
		Description: description,
	}
	
	return migration, nil
}

// parseSQL 解析SQL内容，分离UP和DOWN部分
func (l *MigrationLoader) parseSQL(content string) (upSQL, downSQL, description string, err error) {
	lines := strings.Split(content, "\n")
	var currentSection string
	var upLines, downLines, descLines []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 检查特殊注释标记
		switch {
		case strings.HasPrefix(trimmed, "-- +migrate Up"):
			currentSection = "up"
			continue
		case strings.HasPrefix(trimmed, "-- +migrate Down"):
			currentSection = "down"
			continue
		case strings.HasPrefix(trimmed, "-- Description:"):
			desc := strings.TrimPrefix(trimmed, "-- Description:")
			descLines = append(descLines, strings.TrimSpace(desc))
			continue
		}
		
		// 根据当前section添加到相应的slice
		switch currentSection {
		case "up":
			upLines = append(upLines, line)
		case "down":
			downLines = append(downLines, line)
		default:
			// 如果没有明确的section标记，默认为up
			if currentSection == "" {
				currentSection = "up"
			}
			upLines = append(upLines, line)
		}
	}
	
	upSQL = strings.TrimSpace(strings.Join(upLines, "\n"))
	downSQL = strings.TrimSpace(strings.Join(downLines, "\n"))
	description = strings.Join(descLines, " ")
	
	if upSQL == "" {
		return "", "", "", fmt.Errorf("migration file contains no UP SQL")
	}
	
	return upSQL, downSQL, description, nil
}

// ValidateMigrationFiles 验证迁移文件
func (l *MigrationLoader) ValidateMigrationFiles(migrations []MigrationFile) error {
	versionSet := make(map[string]bool)
	
	for _, migration := range migrations {
		// 检查版本号重复
		if versionSet[migration.Version] {
			return fmt.Errorf("duplicate migration version: %s", migration.Version)
		}
		versionSet[migration.Version] = true
		
		// 验证版本号格式
		if !regexp.MustCompile(`^\d{8}_\d{6}$`).MatchString(migration.Version) {
			return fmt.Errorf("invalid version format: %s (expected: YYYYMMDD_HHMMSS)", migration.Version)
		}
		
		// 验证名称
		if strings.TrimSpace(migration.Name) == "" {
			return fmt.Errorf("migration %s has empty name", migration.Version)
		}
		
		// 验证SQL内容
		if strings.TrimSpace(migration.UpSQL) == "" {
			return fmt.Errorf("migration %s has empty UP SQL", migration.Version)
		}
	}
	
	l.logger.Info("Migration files validation passed", zap.Int("count", len(migrations)))
	return nil
}