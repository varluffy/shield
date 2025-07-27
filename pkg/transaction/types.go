// Package transaction provides transaction management functionality.
// It supports various transaction propagation behaviors and nested transactions.
package transaction

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TransactionPropagation 事务传播行为
type TransactionPropagation int

const (
	// PROPAGATION_REQUIRED 如果存在事务则加入，否则创建新事务（默认）
	PROPAGATION_REQUIRED TransactionPropagation = iota

	// PROPAGATION_REQUIRES_NEW 总是创建新事务，如果存在事务则挂起
	PROPAGATION_REQUIRES_NEW

	// PROPAGATION_NESTED 如果存在事务则创建嵌套事务（Savepoint）
	PROPAGATION_NESTED
)

// String 返回传播行为的字符串表示
func (p TransactionPropagation) String() string {
	switch p {
	case PROPAGATION_REQUIRED:
		return "REQUIRED"
	case PROPAGATION_REQUIRES_NEW:
		return "REQUIRES_NEW"
	case PROPAGATION_NESTED:
		return "NESTED"
	default:
		return "UNKNOWN"
	}
}

// TransactionFunc 事务执行函数类型
type TransactionFunc func(txCtx context.Context) error

// TransactionInfo 事务状态信息
type TransactionInfo struct {
	DB              *gorm.DB                 // 事务数据库连接
	IsTransaction   bool                     // 是否在事务中
	SavepointStack  []string                 // Savepoint栈
	StartTime       time.Time                // 事务开始时间
	PropagationPath []TransactionPropagation // 传播路径（用于调试）
}

// TransactionManager 事务管理器接口
type TransactionManager interface {
	// ExecuteInTransaction 在事务中执行（默认REQUIRED传播）
	ExecuteInTransaction(ctx context.Context, fn TransactionFunc) error

	// ExecuteWithPropagation 使用指定传播行为执行事务
	ExecuteWithPropagation(ctx context.Context, propagation TransactionPropagation, fn TransactionFunc) error

	// GetDB 获取当前上下文的数据库连接（事务或常规）
	GetDB(ctx context.Context) *gorm.DB

	// IsInTransaction 检查当前是否在事务中
	IsInTransaction(ctx context.Context) bool

	// GetTransactionInfo 获取事务信息（用于调试）
	GetTransactionInfo(ctx context.Context) *TransactionInfo
}

// contextKey 用于Context键的类型安全
type contextKey string

const (
	txInfoKey contextKey = "transaction_info"
)
