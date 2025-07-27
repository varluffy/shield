package transaction

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionManagerImpl TransactionManager的实现
type TransactionManagerImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTransactionManager 创建新的事务管理器
func NewTransactionManager(db *gorm.DB, logger *zap.Logger) TransactionManager {
	return &TransactionManagerImpl{
		db:     db,
		logger: logger,
	}
}

// ExecuteInTransaction 在事务中执行（默认REQUIRED传播）
func (tm *TransactionManagerImpl) ExecuteInTransaction(ctx context.Context, fn TransactionFunc) error {
	return tm.ExecuteWithPropagation(ctx, PROPAGATION_REQUIRED, fn)
}

// ExecuteWithPropagation 使用指定传播行为执行事务
func (tm *TransactionManagerImpl) ExecuteWithPropagation(
	ctx context.Context,
	propagation TransactionPropagation,
	fn TransactionFunc,
) error {
	existingTx := tm.GetTransactionInfo(ctx)

	tm.logger.Debug("Executing transaction with propagation",
		zap.String("propagation", propagation.String()),
		zap.Bool("has_existing_tx", existingTx != nil),
	)

	switch propagation {
	case PROPAGATION_REQUIRED:
		return tm.handleRequired(ctx, existingTx, fn)
	case PROPAGATION_REQUIRES_NEW:
		return tm.handleRequiresNew(ctx, existingTx, fn)
	case PROPAGATION_NESTED:
		return tm.handleNested(ctx, existingTx, fn)
	default:
		return fmt.Errorf("unsupported transaction propagation: %v", propagation)
	}
}

// GetDB 获取当前上下文的数据库连接
func (tm *TransactionManagerImpl) GetDB(ctx context.Context) *gorm.DB {
	if txInfo := tm.GetTransactionInfo(ctx); txInfo != nil {
		return txInfo.DB
	}
	return tm.db
}

// IsInTransaction 检查当前是否在事务中
func (tm *TransactionManagerImpl) IsInTransaction(ctx context.Context) bool {
	txInfo := tm.GetTransactionInfo(ctx)
	return txInfo != nil && txInfo.IsTransaction
}

// GetTransactionInfo 获取事务信息
func (tm *TransactionManagerImpl) GetTransactionInfo(ctx context.Context) *TransactionInfo {
	if txInfo, ok := ctx.Value(txInfoKey).(*TransactionInfo); ok {
		return txInfo
	}
	return nil
}

// handleRequired 处理REQUIRED传播行为
func (tm *TransactionManagerImpl) handleRequired(
	ctx context.Context,
	existingTx *TransactionInfo,
	fn TransactionFunc,
) error {
	if existingTx != nil && existingTx.IsTransaction {
		// 加入现有事务
		tm.logger.Debug("Joining existing transaction")
		newTxInfo := tm.copyTransactionInfo(existingTx)
		newTxInfo.PropagationPath = append(newTxInfo.PropagationPath, PROPAGATION_REQUIRED)
		newCtx := tm.withTransactionInfo(ctx, newTxInfo)
		return fn(newCtx)
	}

	// 创建新事务
	tm.logger.Debug("Creating new transaction for REQUIRED propagation")
	return tm.createNewTransaction(ctx, PROPAGATION_REQUIRED, fn)
}

// handleRequiresNew 处理REQUIRES_NEW传播行为
func (tm *TransactionManagerImpl) handleRequiresNew(
	ctx context.Context,
	existingTx *TransactionInfo,
	fn TransactionFunc,
) error {
	// 总是创建新事务，不管是否已有事务
	tm.logger.Debug("Creating new transaction for REQUIRES_NEW propagation",
		zap.Bool("suspending_existing", existingTx != nil),
	)

	// 创建独立的新事务（不继承现有事务的Context）
	return tm.createNewTransaction(context.Background(), PROPAGATION_REQUIRES_NEW, fn)
}

// handleNested 处理NESTED传播行为
func (tm *TransactionManagerImpl) handleNested(
	ctx context.Context,
	existingTx *TransactionInfo,
	fn TransactionFunc,
) error {
	if existingTx == nil || !existingTx.IsTransaction {
		// 没有现有事务，创建新事务
		tm.logger.Debug("No existing transaction, creating new one for NESTED")
		return tm.createNewTransaction(ctx, PROPAGATION_NESTED, fn)
	}

	// 在现有事务中创建Savepoint
	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())

	tm.logger.Debug("Creating savepoint for nested transaction",
		zap.String("savepoint", savepointName),
	)

	// 创建Savepoint
	if err := existingTx.DB.SavePoint(savepointName).Error; err != nil {
		tm.logger.Error("Failed to create savepoint",
			zap.String("savepoint", savepointName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create savepoint %s: %w", savepointName, err)
	}

	// 更新Context，添加Savepoint信息
	newTxInfo := tm.copyTransactionInfo(existingTx)
	newTxInfo.SavepointStack = append(newTxInfo.SavepointStack, savepointName)
	newTxInfo.PropagationPath = append(newTxInfo.PropagationPath, PROPAGATION_NESTED)
	newCtx := tm.withTransactionInfo(ctx, newTxInfo)

	// 执行业务逻辑
	err := fn(newCtx)

	if err != nil {
		// 回滚到Savepoint
		tm.logger.Debug("Rolling back to savepoint due to error",
			zap.String("savepoint", savepointName),
			zap.Error(err),
		)

		if rollbackErr := existingTx.DB.RollbackTo(savepointName).Error; rollbackErr != nil {
			tm.logger.Error("Failed to rollback to savepoint",
				zap.String("savepoint", savepointName),
				zap.Error(rollbackErr),
			)
			return fmt.Errorf("nested transaction failed and rollback failed: %w, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	// 成功，释放Savepoint
	if err := existingTx.DB.Exec("RELEASE SAVEPOINT " + savepointName).Error; err != nil {
		tm.logger.Warn("Failed to release savepoint",
			zap.String("savepoint", savepointName),
			zap.Error(err),
		)
		// 不返回错误，因为业务逻辑已经成功
	}

	tm.logger.Debug("Nested transaction completed successfully",
		zap.String("savepoint", savepointName),
	)

	return nil
}

// createNewTransaction 创建新事务
func (tm *TransactionManagerImpl) createNewTransaction(
	ctx context.Context,
	propagation TransactionPropagation,
	fn TransactionFunc,
) error {
	// 开始事务
	tx := tm.db.Begin()
	if tx.Error != nil {
		tm.logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 创建事务信息
	txInfo := &TransactionInfo{
		DB:              tx,
		IsTransaction:   true,
		SavepointStack:  make([]string, 0),
		StartTime:       time.Now(),
		PropagationPath: []TransactionPropagation{propagation},
	}

	// 将事务信息放入Context
	txCtx := tm.withTransactionInfo(ctx, txInfo)

	tm.logger.Debug("Transaction started",
		zap.String("propagation", propagation.String()),
		zap.Time("start_time", txInfo.StartTime),
	)

	// 执行业务逻辑
	err := fn(txCtx)

	if err != nil {
		// 回滚事务
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			tm.logger.Error("Failed to rollback transaction",
				zap.Error(rollbackErr),
				zap.Error(err),
			)
			return fmt.Errorf("transaction failed and rollback failed: %w, rollback error: %v", err, rollbackErr)
		}

		tm.logger.Debug("Transaction rolled back due to error",
			zap.Error(err),
			zap.Duration("duration", time.Since(txInfo.StartTime)),
		)
		return err
	}

	// 提交事务
	if commitErr := tx.Commit().Error; commitErr != nil {
		tm.logger.Error("Failed to commit transaction", zap.Error(commitErr))
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	tm.logger.Debug("Transaction committed successfully",
		zap.String("propagation", propagation.String()),
		zap.Duration("duration", time.Since(txInfo.StartTime)),
	)

	return nil
}

// withTransactionInfo 将事务信息放入Context
func (tm *TransactionManagerImpl) withTransactionInfo(ctx context.Context, txInfo *TransactionInfo) context.Context {
	return context.WithValue(ctx, txInfoKey, txInfo)
}

// copyTransactionInfo 复制事务信息（用于嵌套事务）
func (tm *TransactionManagerImpl) copyTransactionInfo(txInfo *TransactionInfo) *TransactionInfo {
	newInfo := &TransactionInfo{
		DB:            txInfo.DB,
		IsTransaction: txInfo.IsTransaction,
		StartTime:     txInfo.StartTime,
	}

	// 复制SavepointStack
	if len(txInfo.SavepointStack) > 0 {
		newInfo.SavepointStack = make([]string, len(txInfo.SavepointStack))
		copy(newInfo.SavepointStack, txInfo.SavepointStack)
	} else {
		newInfo.SavepointStack = make([]string, 0)
	}

	// 复制PropagationPath
	if len(txInfo.PropagationPath) > 0 {
		newInfo.PropagationPath = make([]TransactionPropagation, len(txInfo.PropagationPath))
		copy(newInfo.PropagationPath, txInfo.PropagationPath)
	} else {
		newInfo.PropagationPath = make([]TransactionPropagation, 0)
	}

	return newInfo
}
