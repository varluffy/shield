package transaction

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseRepository 提供事务支持的Repository基类
type BaseRepository struct {
	db        *gorm.DB
	txManager TransactionManager
	logger    *zap.Logger
}

// NewBaseRepository 创建基础Repository
func NewBaseRepository(db *gorm.DB, txManager TransactionManager, logger *zap.Logger) *BaseRepository {
	return &BaseRepository{
		db:        db,
		txManager: txManager,
		logger:    logger,
	}
}

// GetDB 获取当前上下文的数据库连接
// 优先返回事务中的DB，否则返回常规DB
func (r *BaseRepository) GetDB(ctx context.Context) *gorm.DB {
	return r.txManager.GetDB(ctx)
}

// IsInTransaction 检查当前是否在事务中
func (r *BaseRepository) IsInTransaction(ctx context.Context) bool {
	return r.txManager.IsInTransaction(ctx)
}

// ExecuteInTransaction 在Repository层执行事务
// 注意：通常事务应该在Service层管理，但某些情况下Repository也可以使用
func (r *BaseRepository) ExecuteInTransaction(ctx context.Context, fn TransactionFunc) error {
	return r.txManager.ExecuteInTransaction(ctx, fn)
}

// GetTransactionInfo 获取当前事务信息（用于调试）
func (r *BaseRepository) GetTransactionInfo(ctx context.Context) *TransactionInfo {
	return r.txManager.GetTransactionInfo(ctx)
}

// LogTransactionState 记录事务状态（用于调试）
func (r *BaseRepository) LogTransactionState(ctx context.Context, operation string) {
	if r.logger != nil {
		txInfo := r.GetTransactionInfo(ctx)
		if txInfo != nil {
			r.logger.Debug("Repository operation with transaction",
				zap.String("operation", operation),
				zap.Bool("in_transaction", txInfo.IsTransaction),
				zap.Int("savepoint_count", len(txInfo.SavepointStack)),
				zap.Strings("propagation_path", r.propagationPathToStrings(txInfo.PropagationPath)),
			)
		} else {
			r.logger.Debug("Repository operation without transaction",
				zap.String("operation", operation),
			)
		}
	}
}

// propagationPathToStrings 将传播路径转换为字符串数组
func (r *BaseRepository) propagationPathToStrings(path []TransactionPropagation) []string {
	result := make([]string, len(path))
	for i, p := range path {
		result[i] = p.String()
	}
	return result
}
