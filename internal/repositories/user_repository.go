// Package repositories contains data access layer implementations.
// It provides repository pattern implementations for database operations.
package repositories

import (
	"context"
	"fmt"

	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/transaction"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate mockgen -source=user_repository.go -destination=mocks/user_repository_mock.go

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint64) (*models.User, error)
	GetByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByEmailAndTenant(ctx context.Context, email string, tenantID uint64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint64) error
	DeleteByUUID(ctx context.Context, uuid string) error
	List(ctx context.Context, filter dto.UserFilter) ([]*models.User, int64, error)
	ListByTenant(ctx context.Context, tenantID uint64, filter dto.UserFilter) ([]*models.User, int64, error)
	Transaction(ctx context.Context, fn func(*gorm.DB) error) error
}

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	*transaction.BaseRepository
	logger *logger.Logger
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB, txManager transaction.TransactionManager, logger *logger.Logger) UserRepository {
	return &UserRepositoryImpl{
		BaseRepository: transaction.NewBaseRepository(db, txManager, logger.Logger),
		logger:         logger,
	}
}

// Create 创建用户
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	r.LogTransactionState(ctx, "Create User")
	r.logger.DebugWithTrace(ctx, "Creating user in database",
		zap.String("email", user.Email),
		zap.String("uuid", user.UUID),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Create(user).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to create user in database",
			zap.Error(err),
			zap.String("email", user.Email),
			zap.String("uuid", user.UUID),
		)
		return err
	}

	r.logger.DebugWithTrace(ctx, "User created in database",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
	)

	return nil
}

// GetByID 根据ID获取用户（内部调用）
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	var user models.User

	r.logger.DebugWithTrace(ctx, "Getting user by ID",
		zap.Uint64("user_id", id),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "User not found by ID",
				zap.Uint64("user_id", id),
			)
			return nil, ErrUserNotFound
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get user by ID",
			zap.Error(err),
			zap.Uint64("user_id", id),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByUUID 根据UUID获取用户（对外接口调用）
func (r *UserRepositoryImpl) GetByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User

	r.logger.DebugWithTrace(ctx, "Getting user by UUID",
		zap.String("user_uuid", uuid),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "User not found by UUID",
				zap.String("user_uuid", uuid),
			)
			return nil, ErrUserNotFound
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get user by UUID",
			zap.Error(err),
			zap.String("user_uuid", uuid),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail 根据邮箱获取用户（兼容旧版本）
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	r.logger.DebugWithTrace(ctx, "Getting user by email",
		zap.String("email", email),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "User not found by email",
				zap.String("email", email),
			)
			return nil, ErrUserNotFound
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get user by email",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmailAndTenant 根据邮箱和租户ID获取用户
func (r *UserRepositoryImpl) GetByEmailAndTenant(ctx context.Context, email string, tenantID uint64) (*models.User, error) {
	var user models.User

	r.logger.DebugWithTrace(ctx, "Getting user by email and tenant",
		zap.String("email", email),
		zap.Uint64("tenant_id", tenantID),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "User not found by email and tenant",
				zap.String("email", email),
				zap.Uint64("tenant_id", tenantID),
			)
			return nil, ErrUserNotFound
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get user by email and tenant",
			zap.Error(err),
			zap.String("email", email),
			zap.Uint64("tenant_id", tenantID),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	r.LogTransactionState(ctx, "Update User")
	r.logger.DebugWithTrace(ctx, "Updating user in database",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Save(user).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to update user in database",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return err
	}

	r.logger.DebugWithTrace(ctx, "User updated in database",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
	)

	return nil
}

// Delete 删除用户（软删除）- 通过ID
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint64) error {
	r.LogTransactionState(ctx, "Delete User by ID")
	r.logger.DebugWithTrace(ctx, "Deleting user from database",
		zap.Uint64("user_id", id),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Delete(&models.User{}, id).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to delete user from database",
			zap.Error(err),
			zap.Uint64("user_id", id),
		)
		return err
	}

	r.logger.DebugWithTrace(ctx, "User deleted from database",
		zap.Uint64("user_id", id),
	)

	return nil
}

// DeleteByUUID 删除用户（软删除）- 通过UUID
func (r *UserRepositoryImpl) DeleteByUUID(ctx context.Context, uuid string) error {
	r.LogTransactionState(ctx, "Delete User by UUID")
	r.logger.DebugWithTrace(ctx, "Deleting user from database",
		zap.String("user_uuid", uuid),
		zap.Bool("in_transaction", r.IsInTransaction(ctx)),
	)

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&models.User{}).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to delete user from database",
			zap.Error(err),
			zap.String("user_uuid", uuid),
		)
		return err
	}

	r.logger.DebugWithTrace(ctx, "User deleted from database",
		zap.String("user_uuid", uuid),
	)

	return nil
}

// List 获取用户列表（兼容旧版本）
func (r *UserRepositoryImpl) List(ctx context.Context, filter dto.UserFilter) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	r.logger.DebugWithTrace(ctx, "Getting user list",
		zap.Int("page", filter.Page),
		zap.Int("limit", filter.Limit),
	)

	db := r.GetDB(ctx)
	query := db.WithContext(ctx).Model(&models.User{})

	// 添加筛选条件
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Active != nil {
		query = query.Where("active = ?", *filter.Active)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to count users",
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页和排序
	offset := (filter.Page - 1) * filter.Limit
	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.OrderBy, filter.OrderDir)
	}

	err := query.Order(orderBy).Offset(offset).Limit(filter.Limit).Find(&users).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to list users",
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Got user list",
		zap.Int("count", len(users)),
		zap.Int64("total", total),
	)

	return users, total, nil
}

// ListByTenant 获取指定租户的用户列表
func (r *UserRepositoryImpl) ListByTenant(ctx context.Context, tenantID uint64, filter dto.UserFilter) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	r.logger.DebugWithTrace(ctx, "Getting user list by tenant",
		zap.Uint64("tenant_id", tenantID),
		zap.Int("page", filter.Page),
		zap.Int("limit", filter.Limit),
	)

	db := r.GetDB(ctx)
	query := db.WithContext(ctx).Model(&models.User{}).Where("tenant_id = ?", tenantID)

	// 添加筛选条件
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Active != nil {
		query = query.Where("active = ?", *filter.Active)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to count users by tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
		)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页和排序
	offset := (filter.Page - 1) * filter.Limit
	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.OrderBy, filter.OrderDir)
	}

	err := query.Order(orderBy).Offset(offset).Limit(filter.Limit).Find(&users).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to list users by tenant",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
		)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Got user list by tenant",
		zap.Uint64("tenant_id", tenantID),
		zap.Int("count", len(users)),
		zap.Int64("total", total),
	)

	return users, total, nil
}

// Transaction 执行事务（为了向后兼容保留，推荐使用TransactionManager）
func (r *UserRepositoryImpl) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		db := r.GetDB(txCtx)
		return fn(db)
	})
}
