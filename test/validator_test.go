package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/pkg/validator"
)

// TestMultiLanguageValidation 测试多语言验证功能
func TestMultiLanguageValidation(t *testing.T) {
	tests := []struct {
		name           string
		language       string
		requestBody    dto.CreateUserRequest
		expectedErrors []string
		description    string
	}{
		{
			name:     "中文验证错误",
			language: "zh",
			requestBody: dto.CreateUserRequest{
				Name:     "",              // 缺少必填字段
				Email:    "invalid-email", // 无效邮箱
				Password: "123",           // 密码太短
			},
			expectedErrors: []string{
				"姓名为必填字段",
				"邮箱必须是有效的邮箱地址",
				"密码长度必须至少为8个字符",
			},
			description: "验证中文错误信息是否正确显示",
		},
		{
			name:     "英文验证错误",
			language: "en",
			requestBody: dto.CreateUserRequest{
				Name:     "",
				Email:    "invalid-email",
				Password: "123",
			},
			expectedErrors: []string{
				"姓名 is required",
				"邮箱 must be a valid email address",
				"密码 must be at least 8 characters long",
			},
			description: "验证英文错误信息是否正确显示",
		},
		{
			name:     "中文字段长度验证",
			language: "zh",
			requestBody: dto.CreateUserRequest{
				Name:     "A", // 太短
				Email:    "test@example.com",
				Password: "verylongpasswordthatexceedsthemaximumlengthallowedforpasswordfieldswhichissetto128charactersmaximumsoletsmakeitevenlonger", // 太长
			},
			expectedErrors: []string{
				"姓名长度必须至少为2个字符",
			},
			description: "验证字段长度限制的中文错误信息",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建指定语言的验证器
			v, err := validator.NewValidator(tt.language)
			assert.NoError(t, err)

			// 执行验证
			errorMessages, validationErr := v.ValidateAndTranslate(tt.requestBody)
			assert.Error(t, validationErr) // 应该有验证错误
			assert.NotEmpty(t, errorMessages)

			// 验证错误信息包含预期的验证错误
			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, actualMsg := range errorMessages {
					if actualMsg == expectedError {
						found = true
						break
					}
				}
				assert.True(t, found,
					"Expected error message '%s' not found in: %v",
					expectedError, errorMessages)
			}

			t.Logf("测试 %s: %s", tt.name, tt.description)
			t.Logf("验证错误信息: %v", errorMessages)
		})
	}
}

// TestValidatorDirectUsage 测试直接使用验证器
func TestValidatorDirectUsage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		data     interface{}
		expected []string
	}{
		{
			name:     "中文验证-创建用户",
			language: "zh",
			data: dto.CreateUserRequest{
				Name:     "",
				Email:    "invalid",
				Password: "short",
			},
			expected: []string{"姓名为必填字段", "邮箱必须是有效的邮箱地址", "密码长度必须至少为8个字符"},
		},
		{
			name:     "英文验证-创建用户",
			language: "en",
			data: dto.CreateUserRequest{
				Name:     "",
				Email:    "invalid",
				Password: "short",
			},
			expected: []string{"姓名 is required", "邮箱 must be a valid email address", "密码 must be at least 8 characters long"},
		},
		{
			name:     "中文验证-登录",
			language: "zh",
			data: dto.LoginRequest{
				Email:     "",
				Password:  "",
				CaptchaID: "",
				Answer:    "",
			},
			expected: []string{"email为必填字段", "password为必填字段", "captcha_id为必填字段", "answer为必填字段"},
		},
		{
			name:     "英文验证-登录",
			language: "en",
			data: dto.LoginRequest{
				Email:     "",
				Password:  "",
				CaptchaID: "",
				Answer:    "",
			},
			expected: []string{"email is required", "password is required", "captcha_id is required", "answer is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建验证器
			v, err := validator.NewValidator(tt.language)
			assert.NoError(t, err)

			// 执行验证
			errorMessages, validationErr := v.ValidateAndTranslate(tt.data)
			assert.Error(t, validationErr) // 应该有验证错误
			assert.NotEmpty(t, errorMessages)

			// 验证错误信息
			for _, expectedMsg := range tt.expected {
				found := false
				for _, actualMsg := range errorMessages {
					if actualMsg == expectedMsg {
						found = true
						break
					}
				}
				assert.True(t, found,
					"Expected error message '%s' not found in: %v",
					expectedMsg, errorMessages)
			}

			t.Logf("语言: %s, 错误信息: %v", tt.language, errorMessages)
		})
	}
}

// TestLanguageSwitching 测试语言切换
func TestLanguageSwitching(t *testing.T) {
	// 初始化中文验证器
	err := validator.InitGlobalValidator("zh")
	assert.NoError(t, err)

	v := validator.GetGlobalValidator()
	assert.Equal(t, "zh", v.GetLanguage())

	// 测试中文验证
	testData := dto.CreateUserRequest{
		Name:     "",
		Email:    "invalid",
		Password: "short",
	}

	errorMessages, _ := v.ValidateAndTranslate(testData)
	assert.Contains(t, errorMessages[0], "为必填字段") // 中文错误信息

	// 切换到英文
	err = v.SetLanguage("en")
	assert.NoError(t, err)
	assert.Equal(t, "en", v.GetLanguage())

	// 测试英文验证
	errorMessages, _ = v.ValidateAndTranslate(testData)
	assert.Contains(t, errorMessages[0], "is required") // 英文错误信息

	t.Logf("语言切换测试成功")
}

// TestUnsupportedLanguage 测试不支持的语言
func TestUnsupportedLanguage(t *testing.T) {
	// 尝试使用不支持的语言，应该回退到默认语言
	v, err := validator.NewValidator("fr") // 法语不支持
	assert.NoError(t, err)
	assert.Equal(t, validator.DefaultLanguage, v.GetLanguage()) // 应该回退到默认语言

	t.Logf("不支持的语言测试成功，回退到默认语言: %s", v.GetLanguage())
}

// TestValidatorWithSuccess 测试验证成功的情况
func TestValidatorWithSuccess(t *testing.T) {
	tests := []struct {
		name     string
		language string
		data     interface{}
	}{
		{
			name:     "中文验证-有效用户数据",
			language: "zh",
			data: dto.CreateUserRequest{
				Name:     "张三",
				Email:    "zhangsan@example.com",
				Password: "password123",
			},
		},
		{
			name:     "英文验证-有效用户数据",
			language: "en",
			data: dto.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
		},
		{
			name:     "中文验证-有效登录数据",
			language: "zh",
			data: dto.LoginRequest{
				Email:     "user@example.com",
				Password:  "password123",
				CaptchaID: "test-captcha-id",
				Answer:    "1234",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建验证器
			v, err := validator.NewValidator(tt.language)
			assert.NoError(t, err)

			// 执行验证
			errorMessages, validationErr := v.ValidateAndTranslate(tt.data)
			assert.NoError(t, validationErr) // 不应该有验证错误
			assert.Empty(t, errorMessages)   // 错误信息应该为空

			t.Logf("验证成功: %s", tt.name)
		})
	}
}

// TestValidatorFieldSpecificErrors 测试特定字段错误
func TestValidatorFieldSpecificErrors(t *testing.T) {
	v, err := validator.NewValidator("zh")
	assert.NoError(t, err)

	// 测试邮箱格式错误
	testData := dto.CreateUserRequest{
		Name:     "测试用户",
		Email:    "not-an-email",
		Password: "password123",
	}

	errorMessages, validationErr := v.ValidateAndTranslate(testData)
	assert.Error(t, validationErr)
	assert.NotEmpty(t, errorMessages)

	// 应该只有邮箱错误
	assert.Len(t, errorMessages, 1)
	assert.Contains(t, errorMessages[0], "邮箱必须是有效的邮箱地址")

	t.Logf("特定字段错误测试成功: %v", errorMessages)
}

// mockUserService 简化的模拟用户服务（用于未来可能的集成测试）
type mockUserService struct{}

func (m *mockUserService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserService) ListUsers(ctx context.Context, filter dto.UserFilter) (*dto.UserListResponse, error) {
	return nil, nil
}

func (m *mockUserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	return nil, nil
}

// 添加缺失的事务管理方法
func (m *mockUserService) CreateUsersBatch(ctx context.Context, users []dto.CreateUserRequest) ([]*dto.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) TransferUserRole(ctx context.Context, fromUserID, toUserID uint) error {
	return nil
}
