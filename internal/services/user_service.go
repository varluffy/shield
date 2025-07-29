// Package services contains business logic implementations.
// It provides service layer that orchestrates business operations and rules.
package services

import (
	"context"
	"fmt"
	"math"

	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/auth"
	"github.com/varluffy/shield/pkg/captcha"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/transaction"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=user_service.go -destination=mocks/user_service_mock.go

// UserService 用户服务接口
type UserService interface {
	// 对外接口 - 使用UUID
	CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error)
	UpdateUserByUUID(ctx context.Context, uuid string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUserByUUID(ctx context.Context, uuid string) error
	ListUsers(ctx context.Context, filter dto.UserFilter) (*dto.UserListResponse, error)

	// 内部接口 - 使用ID（为了兼容和内部调用）
	GetUserByID(ctx context.Context, id uint64) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id uint64, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uint64) error

	// 认证相关
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)

	// 事务管理演示方法
	CreateUsersBatch(ctx context.Context, users []dto.CreateUserRequest) ([]*dto.UserResponse, error)
	TransferUserRole(ctx context.Context, fromUserUUID, toUserUUID string) error
}

// UserServiceImpl 用户服务实现
type UserServiceImpl struct {
	userRepo       repositories.UserRepository
	logger         *logger.Logger
	txManager      transaction.TransactionManager
	jwtService     auth.JWTService
	captchaService captcha.CaptchaService
	config         *config.Config
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo repositories.UserRepository,
	logger *logger.Logger,
	txManager transaction.TransactionManager,
	jwtService auth.JWTService,
	captchaService captcha.CaptchaService,
	config *config.Config,
) UserService {
	return &UserServiceImpl{
		userRepo:       userRepo,
		logger:         logger,
		txManager:      txManager,
		jwtService:     jwtService,
		captchaService: captchaService,
		config:         config,
	}
}

// CreateUser 创建用户
func (s *UserServiceImpl) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	s.logger.InfoWithTrace(ctx, "Creating user",
		zap.String("name", req.Name),
		zap.String("email", req.Email),
	)

	// 从上下文中获取当前用户的tenant_id
	tenantID, exists := ctx.Value("tenant_id").(uint64)
	if !exists || tenantID == 0 {
		s.logger.ErrorWithTrace(ctx, "Tenant ID not found in context")
		return nil, errors.ErrInternalError("tenant context not found")
	}

	// 验证用户数据
	if err := s.validateCreateUserRequest(req); err != nil {
		s.logger.WarnWithTrace(ctx, "Invalid user data",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, err
	}

	// 检查邮箱是否已存在（在同一租户内）
	existingUser, err := s.userRepo.GetByEmailAndTenant(ctx, req.Email, tenantID)
	if err != nil && err != repositories.ErrUserNotFound {
		s.logger.ErrorWithTrace(ctx, "Failed to check existing user",
			zap.Error(err),
			zap.String("email", req.Email),
			zap.Uint64("tenant_id", tenantID),
		)
		return nil, errors.ErrInternalError("failed to check existing user")
	}
	if existingUser != nil {
		s.logger.WarnWithTrace(ctx, "User email already exists in tenant",
			zap.String("email", req.Email),
			zap.Uint64("tenant_id", tenantID),
		)
		return nil, errors.ErrUserEmailExists()
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to hash password",
			zap.Error(err),
		)
		return nil, errors.ErrInternalError("failed to hash password")
	}

	// 创建用户模型
	user := &models.User{
		TenantModel: models.TenantModel{TenantID: tenantID}, // 设置tenant_id
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		Status:      "active",
	}

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to create user",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, errors.ErrInternalError("failed to create user")
	}

	s.logger.InfoWithTrace(ctx, "User created successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
		zap.String("email", user.Email),
		zap.Uint64("tenant_id", user.TenantID),
	)

	return s.modelToResponse(user), nil
}

// GetUserByUUID 根据UUID获取用户（对外接口）
func (s *UserServiceImpl) GetUserByUUID(ctx context.Context, uuid string) (*dto.UserResponse, error) {
	s.logger.DebugWithTrace(ctx, "Getting user by UUID",
		zap.String("user_uuid", uuid),
	)

	user, err := s.userRepo.GetByUUID(ctx, uuid)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user by UUID",
			zap.Error(err),
			zap.String("user_uuid", uuid),
		)
		return nil, err
	}

	return s.modelToResponse(user), nil
}

// GetUserByID 根据ID获取用户（内部调用）
func (s *UserServiceImpl) GetUserByID(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	s.logger.DebugWithTrace(ctx, "Getting user by ID",
		zap.Uint64("user_id", id),
	)

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user by ID",
			zap.Error(err),
			zap.Uint64("user_id", id),
		)
		return nil, err
	}

	return s.modelToResponse(user), nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	s.logger.DebugWithTrace(ctx, "Getting user by email",
		zap.String("email", email),
	)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user by email",
			zap.Error(err),
			zap.String("email", email),
		)
		return nil, err
	}

	return s.modelToResponse(user), nil
}

// UpdateUserByUUID 更新用户（对外接口）
func (s *UserServiceImpl) UpdateUserByUUID(ctx context.Context, uuid string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	s.logger.InfoWithTrace(ctx, "Updating user by UUID",
		zap.String("user_uuid", uuid),
	)

	// 获取现有用户
	user, err := s.userRepo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return s.updateUserInternal(ctx, user, req)
}

// UpdateUser 更新用户（内部调用）
func (s *UserServiceImpl) UpdateUser(ctx context.Context, id uint64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	s.logger.InfoWithTrace(ctx, "Updating user by ID",
		zap.Uint64("user_id", id),
	)

	// 获取现有用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.updateUserInternal(ctx, user, req)
}

// updateUserInternal 内部更新用户逻辑
func (s *UserServiceImpl) updateUserInternal(ctx context.Context, user *models.User, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 更新字段
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		// 检查新邮箱是否已被其他用户使用
		existingUser, err := s.userRepo.GetByEmailAndTenant(ctx, req.Email, user.TenantID)
		if err != nil && err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = req.Email
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to update user",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrInternalError("failed to update user")
	}

	s.logger.InfoWithTrace(ctx, "User updated successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
	)

	return s.modelToResponse(user), nil
}

// DeleteUserByUUID 删除用户（对外接口）
func (s *UserServiceImpl) DeleteUserByUUID(ctx context.Context, uuid string) error {
	s.logger.InfoWithTrace(ctx, "Deleting user by UUID",
		zap.String("user_uuid", uuid),
	)

	if err := s.userRepo.DeleteByUUID(ctx, uuid); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to delete user by UUID",
			zap.Error(err),
			zap.String("user_uuid", uuid),
		)
		return err
	}

	s.logger.InfoWithTrace(ctx, "User deleted successfully by UUID",
		zap.String("user_uuid", uuid),
	)

	return nil
}

// DeleteUser 删除用户（内部调用）
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id uint64) error {
	s.logger.InfoWithTrace(ctx, "Deleting user by ID",
		zap.Uint64("user_id", id),
	)

	if err := s.userRepo.Delete(ctx, id); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to delete user by ID",
			zap.Error(err),
			zap.Uint64("user_id", id),
		)
		return err
	}

	s.logger.InfoWithTrace(ctx, "User deleted successfully by ID",
		zap.Uint64("user_id", id),
	)

	return nil
}

// ListUsers 获取用户列表
func (s *UserServiceImpl) ListUsers(ctx context.Context, filter dto.UserFilter) (*dto.UserListResponse, error) {
	s.logger.DebugWithTrace(ctx, "Listing users",
		zap.Int("page", filter.Page),
		zap.Int("limit", filter.Limit),
	)

	// 从上下文获取租户ID
	if tenantID, exists := ctx.Value("tenant_id").(uint64); exists && tenantID > 0 {
		// 使用租户特定的列表查询
		users, total, err := s.userRepo.ListByTenant(ctx, tenantID, filter)
		if err != nil {
			s.logger.ErrorWithTrace(ctx, "Failed to list users by tenant",
				zap.Error(err),
				zap.Uint64("tenant_id", tenantID),
			)
			return nil, errors.ErrInternalError("failed to list users")
		}

		// 转换为响应格式
		userResponses := make([]dto.UserResponse, len(users))
		for i, user := range users {
			userResponses[i] = *s.modelToResponse(user)
		}

		// 计算分页信息
		totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

		return &dto.UserListResponse{
			Users: userResponses,
			Meta: dto.PaginationMeta{
				Page:      filter.Page,
				Limit:     filter.Limit,
				Total:     int(total),
				TotalPage: totalPages,
			},
		}, nil
	}

	// 回退到全局列表（系统管理员）
	users, total, err := s.userRepo.List(ctx, filter)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to list users",
			zap.Error(err),
		)
		return nil, errors.ErrInternalError("failed to list users")
	}

	// 转换为响应格式
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.modelToResponse(user)
	}

	// 计算分页信息
	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &dto.UserListResponse{
		Users: userResponses,
		Meta: dto.PaginationMeta{
			Page:      filter.Page,
			Limit:     filter.Limit,
			Total:     int(total),
			TotalPage: totalPages,
		},
	}, nil
}

// Login 用户登录
func (s *UserServiceImpl) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	s.logger.InfoWithTrace(ctx, "User login attempt",
		zap.String("email", req.Email),
		zap.String("environment", s.config.App.Environment),
	)

	// 根据环境和配置决定验证码策略
	if s.shouldRequireCaptcha() {
		if s.isDevBypass(req.CaptchaID, req.Answer) {
			s.logger.WarnWithTrace(ctx, "Using development captcha bypass", 
				zap.String("environment", s.config.App.Environment),
				zap.String("captcha_id", req.CaptchaID),
			)
		} else {
			// 正常验证码验证
			if err := s.captchaService.VerifyCaptcha(ctx, req.CaptchaID, req.Answer); err != nil {
				s.logger.WarnWithTrace(ctx, "Login failed - invalid captcha",
					zap.String("email", req.Email),
					zap.String("captcha_id", req.CaptchaID),
					zap.Error(err),
				)
				return nil, errors.ErrCaptchaInvalid()
			}
		}
	}

	// 获取用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			s.logger.WarnWithTrace(ctx, "Login failed - user not found",
				zap.String("email", req.Email),
			)
			return nil, errors.ErrInvalidCredentials()
		}
		s.logger.ErrorWithTrace(ctx, "Failed to get user for login",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// 检查用户是否激活
	if user.Status != "active" {
		s.logger.WarnWithTrace(ctx, "Login failed - user inactive",
			zap.String("email", req.Email),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrUserInactive()
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.WarnWithTrace(ctx, "Login failed - invalid password",
			zap.String("email", req.Email),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrInvalidCredentials()
	}

	// 获取租户UUID（为JWT token使用）
	// 暂时使用TenantID转换为字符串，但此时tenant_id在权限检查时需要特殊处理
	tenantUUID := fmt.Sprintf("%d", user.TenantID)

	// 生成JWT token（使用UUID以提高安全性）
	accessToken, err := s.jwtService.GenerateAccessToken(user.UUID, user.Email, tenantUUID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to generate access token",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrInternalError("failed to generate access token")
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.UUID, tenantUUID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to generate refresh token",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrInternalError("failed to generate refresh token")
	}

	s.logger.InfoWithTrace(ctx, "User logged in successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
		zap.String("email", user.Email),
	)

	return &dto.LoginResponse{
		User:         *s.modelToResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1小时
	}, nil
}


// shouldRequireCaptcha 判断是否应该验证验证码
func (s *UserServiceImpl) shouldRequireCaptcha() bool {
	switch s.config.App.Environment {
	case "production":
		return true // 生产环境总是需要验证码
	case "development", "test":
		return s.config.Auth.CaptchaMode != "disabled"
	default:
		return true // 默认安全策略
	}
}

// isDevBypass 判断是否为开发/测试环境绕过
func (s *UserServiceImpl) isDevBypass(captchaID, answer string) bool {
	return (s.config.App.Environment == "development" || s.config.App.Environment == "test") && 
		   captchaID == "dev-bypass" && 
		   answer == s.config.Auth.DevBypassCode
}

// CreateUsersBatch 批量创建用户（事务演示）
func (s *UserServiceImpl) CreateUsersBatch(ctx context.Context, users []dto.CreateUserRequest) ([]*dto.UserResponse, error) {
	s.logger.InfoWithTrace(ctx, "Creating users batch",
		zap.Int("count", len(users)),
	)

	var responses []*dto.UserResponse

	// 使用事务管理器执行批量操作
	err := s.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		for i, req := range users {
			s.logger.DebugWithTrace(txCtx, "Creating user in batch",
				zap.Int("index", i),
				zap.String("email", req.Email),
			)

			// 检查邮箱是否已存在
			existingUser, err := s.userRepo.GetByEmail(txCtx, req.Email)
			if err != nil && err != repositories.ErrUserNotFound {
				return fmt.Errorf("failed to check existing user %s: %w", req.Email, err)
			}
			if existingUser != nil {
				return fmt.Errorf("user with email %s already exists", req.Email)
			}

			// 加密密码
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password for user %s: %w", req.Email, err)
			}

			// 创建用户模型
			user := &models.User{
				Name:     req.Name,
				Email:    req.Email,
				Password: string(hashedPassword),
				Status:   "active",
			}

			// 保存到数据库（自动使用事务中的DB连接）
			if err := s.userRepo.Create(txCtx, user); err != nil {
				return fmt.Errorf("failed to create user %s: %w", req.Email, err)
			}

			responses = append(responses, s.modelToResponse(user))
		}
		return nil
	})
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to create users batch",
			zap.Error(err),
			zap.Int("count", len(users)),
		)
		return nil, err
	}

	s.logger.InfoWithTrace(ctx, "Users batch created successfully",
		zap.Int("count", len(responses)),
	)

	return responses, nil
}

// TransferUserRole 转移用户角色
func (s *UserServiceImpl) TransferUserRole(ctx context.Context, fromUserUUID, toUserUUID string) error {
	s.logger.InfoWithTrace(ctx, "Transferring user role",
		zap.String("from_user_uuid", fromUserUUID),
		zap.String("to_user_uuid", toUserUUID),
	)

	// 这里需要实现具体的角色转移逻辑
	// 暂时返回未实现错误
	return errors.ErrInternalError("role transfer not implemented")
}

// Register 用户注册
func (s *UserServiceImpl) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserResponse, error) {
	s.logger.InfoWithTrace(ctx, "User registration attempt",
		zap.String("email", req.Email),
		zap.String("name", req.Name),
	)

	// 验证验证码
	if err := s.captchaService.VerifyCaptcha(ctx, req.CaptchaID, req.Answer); err != nil {
		s.logger.WarnWithTrace(ctx, "Registration failed - invalid captcha",
			zap.String("email", req.Email),
			zap.String("captcha_id", req.CaptchaID),
			zap.Error(err),
		)
		return nil, errors.ErrCaptchaInvalid()
	}

	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != repositories.ErrUserNotFound {
		s.logger.ErrorWithTrace(ctx, "Failed to check existing user",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		s.logger.WarnWithTrace(ctx, "Registration failed - user already exists",
			zap.String("email", req.Email),
		)
		return nil, errors.ErrUserAlreadyExists()
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to hash password",
			zap.Error(err),
		)
		return nil, errors.ErrInternalError("failed to hash password")
	}

	// 创建用户模型
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Status:   "active",
	}

	// 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to create user",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, errors.ErrInternalError("failed to create user")
	}

	s.logger.InfoWithTrace(ctx, "User registered successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
		zap.String("email", user.Email),
	)

	return s.modelToResponse(user), nil
}

// RefreshToken 刷新访问令牌
func (s *UserServiceImpl) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	s.logger.InfoWithTrace(ctx, "Token refresh attempt")

	// 验证刷新令牌
	claims, err := s.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		s.logger.WarnWithTrace(ctx, "Token refresh failed - invalid refresh token",
			zap.Error(err),
		)
		return nil, errors.ErrInvalidToken()
	}

	// 获取用户信息（确保用户仍然存在且激活）
	// JWT中的UserID是UUID
	user, err := s.userRepo.GetByUUID(ctx, claims.UserID)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			s.logger.WarnWithTrace(ctx, "Token refresh failed - user not found",
				zap.String("user_uuid", claims.UserID),
			)
			return nil, errors.ErrInvalidToken()
		}
		s.logger.ErrorWithTrace(ctx, "Failed to get user for token refresh",
			zap.Error(err),
			zap.String("user_uuid", claims.UserID),
		)
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// 检查用户是否激活
	if user.Status != "active" {
		s.logger.WarnWithTrace(ctx, "Token refresh failed - user inactive",
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrUserInactive()
	}

	// 获取租户UUID（为JWT token使用）
	tenantUUID := fmt.Sprintf("%d", user.TenantID)

	// 生成新的访问令牌（使用UUID）
	newAccessToken, err := s.jwtService.GenerateAccessToken(user.UUID, user.Email, tenantUUID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to generate new access token",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
			zap.String("user_uuid", user.UUID),
		)
		return nil, errors.ErrInternalError("failed to generate access token")
	}

	s.logger.InfoWithTrace(ctx, "Token refreshed successfully",
		zap.Uint64("user_id", user.ID),
		zap.String("user_uuid", user.UUID),
	)

	return &dto.RefreshTokenResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   3600, // 1小时
	}, nil
}

// modelToResponse 将模型转换为响应DTO（对外只暴露UUID，不暴露内部ID）
func (s *UserServiceImpl) modelToResponse(user *models.User) *dto.UserResponse {
	// TODO: 获取租户UUID，暂时使用TenantID转换
	tenantUUID := fmt.Sprintf("%d", user.TenantID)

	return &dto.UserResponse{
		ID:        user.UUID, // 使用UUID作为对外ID
		Name:      user.Name,
		Email:     user.Email,
		Status:    user.Status,
		Active:    user.Status == "active",
		TenantID:  tenantUUID, // 使用租户UUID作为对外Tenant ID
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// validateCreateUserRequest 验证创建用户请求
func (s *UserServiceImpl) validateCreateUserRequest(req dto.CreateUserRequest) error {
	if req.Name == "" {
		return errors.ErrInvalidRequest()
	}
	if req.Email == "" {
		return errors.ErrInvalidRequest()
	}
	if req.Password == "" {
		return errors.ErrInvalidRequest()
	}
	return nil
}
