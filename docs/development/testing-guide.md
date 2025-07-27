# UltraFit 测试指南

本指南详细说明了 UltraFit 项目的测试策略、工具使用和最佳实践，确保代码质量和系统可靠性。

## 🎯 测试策略

### 测试金字塔

```
                E2E Tests
              ────────────────
            /                \
           /  Integration Tests \
          ──────────────────────────
         /                          \
        /         Unit Tests         \
       ──────────────────────────────────
```

**测试层次**：
- **单元测试 (70%)**: 测试单个函数和方法
- **集成测试 (20%)**: 测试组件间的交互
- **端到端测试 (10%)**: 测试完整的用户场景

### 测试原则

1. **快速反馈**: 单元测试运行时间 < 1 秒
2. **独立性**: 测试之间互不依赖
3. **可重复**: 任何环境下结果一致
4. **清晰明确**: 测试名称表达意图
5. **全面覆盖**: 关键业务逻辑 100% 覆盖

## 🛠️ 测试工具与框架

### 核心测试工具

```go
import (
    "testing"                    // Go 标准测试框架
    "github.com/stretchr/testify/assert"  // 断言库
    "github.com/stretchr/testify/require" // 必要条件断言
    "github.com/stretchr/testify/mock"    // Mock 框架
    "github.com/stretchr/testify/suite"   // 测试套件
    "gorm.io/gorm"              // 数据库测试
    "net/http/httptest"         // HTTP 测试
    "github.com/gin-gonic/gin"  // Web 框架测试
)
```

### 项目测试结构

```
shield/
├── test/                       # 集成测试和工具
│   ├── test_helpers.go        # 测试辅助函数
│   ├── permission_system_test.go  # 权限系统集成测试
│   ├── redis_test.go          # Redis 集成测试
│   └── validator_test.go      # 验证器测试
├── internal/
│   ├── handlers/              # Handler 层测试
│   │   └── *_test.go
│   ├── services/              # Service 层测试
│   │   └── *_test.go
│   └── repositories/          # Repository 层测试
│       └── *_test.go
└── pkg/                       # 公共包测试
    └── *_test.go
```

## 🧪 单元测试实践

### 测试命名规范

```go
// 测试函数命名: TestFunction_Scenario_ExpectedBehavior
func TestUserService_CreateUser_ValidInput_Success(t *testing.T) {}
func TestUserService_CreateUser_EmailExists_ReturnsError(t *testing.T) {}
func TestUserService_CreateUser_InvalidEmail_ReturnsValidationError(t *testing.T) {}
```

### 表驱动测试模式

```go
func TestUserService_ValidateUser(t *testing.T) {
    tests := []struct {
        name        string
        user        *models.User
        wantErr     bool
        expectedErr string
    }{
        {
            name: "有效用户信息",
            user: &models.User{
                Name:  "张三",
                Email: "zhangsan@example.com",
                Phone: "+86-13800138000",
            },
            wantErr: false,
        },
        {
            name: "邮箱格式无效",
            user: &models.User{
                Name:  "张三",
                Email: "invalid-email",
                Phone: "+86-13800138000",
            },
            wantErr:     true,
            expectedErr: "邮箱格式不正确",
        },
        {
            name: "手机号为空",
            user: &models.User{
                Name:  "张三",
                Email: "zhangsan@example.com",
                Phone: "",
            },
            wantErr:     true,
            expectedErr: "手机号不能为空",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := &userService{}
            err := service.ValidateUser(tt.user)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Service 层测试示例

```go
func TestUserService_CreateUser(t *testing.T) {
    // 使用 testify/mock 进行模拟
    mockRepo := new(mocks.UserRepository)
    mockLogger := new(mocks.Logger)
    mockTxManager := new(mocks.TransactionManager)
    
    service := &userService{
        userRepo:  mockRepo,
        logger:    mockLogger,
        txManager: mockTxManager,
    }
    
    t.Run("成功创建用户", func(t *testing.T) {
        user := &models.User{
            Name:     "张三",
            Email:    "zhangsan@example.com",
            TenantID: 1,
        }
        
        // 设置 Mock 期望
        mockRepo.On("GetByEmail", mock.Anything, user.Email).
            Return(nil, gorm.ErrRecordNotFound)
        mockRepo.On("Create", mock.Anything, user).
            Return(nil)
        mockLogger.On("InfoWithTrace", mock.Anything, mock.Anything, mock.Anything)
        
        // 执行测试
        err := service.CreateUser(context.Background(), user)
        
        // 验证结果
        assert.NoError(t, err)
        assert.NotEmpty(t, user.ID)
        
        // 验证 Mock 调用
        mockRepo.AssertExpectations(t)
        mockLogger.AssertExpectations(t)
    })
    
    t.Run("邮箱已存在", func(t *testing.T) {
        user := &models.User{
            Email: "existing@example.com",
        }
        
        existingUser := &models.User{
            ID:    "existing-id",
            Email: "existing@example.com",
        }
        
        mockRepo.On("GetByEmail", mock.Anything, user.Email).
            Return(existingUser, nil)
        
        err := service.CreateUser(context.Background(), user)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "邮箱已存在")
        
        mockRepo.AssertExpectations(t)
    })
}
```

### Repository 层测试示例

```go
func TestUserRepository_Create(t *testing.T) {
    // 设置测试数据库
    db, cleanup := test.SetupTestDB(t)
    if db == nil {
        return // 跳过测试
    }
    defer cleanup()
    
    logger, _ := test.NewTestLogger()
    txManager := transaction.NewTransactionManager(db, logger.Logger)
    repo := repositories.NewUserRepository(db, txManager, logger)
    
    t.Run("成功创建用户", func(t *testing.T) {
        user := &models.User{
            TenantID: 1,
            Name:     "测试用户",
            Email:    "test@example.com",
            Password: "hashed_password",
            Status:   "active",
        }
        
        err := repo.Create(context.Background(), user)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, user.ID)
        assert.NotZero(t, user.CreatedAt)
        
        // 验证数据库中的数据
        var savedUser models.User
        err = db.First(&savedUser, "email = ?", user.Email).Error
        assert.NoError(t, err)
        assert.Equal(t, user.Name, savedUser.Name)
        assert.Equal(t, user.Email, savedUser.Email)
    })
    
    t.Run("邮箱重复创建失败", func(t *testing.T) {
        // 先创建一个用户
        user1 := &models.User{
            TenantID: 1,
            Email:    "duplicate@example.com",
            Name:     "用户1",
            Password: "password",
        }
        err := repo.Create(context.Background(), user1)
        require.NoError(t, err)
        
        // 创建相同邮箱的用户（应该失败）
        user2 := &models.User{
            TenantID: 1,
            Email:    "duplicate@example.com", // 重复邮箱
            Name:     "用户2",
            Password: "password",
        }
        
        err = repo.Create(context.Background(), user2)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "duplicate")
    })
}
```

## 🔌 集成测试

### HTTP Handler 集成测试

```go
func TestUserHandler_Integration(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    
    db, cleanup := test.SetupTestDB(t)
    if db == nil {
        return
    }
    defer cleanup()
    
    // 创建测试组件
    testLogger, _ := test.NewTestLogger()
    components := test.NewTestComponents(db, testLogger)
    
    // 设置路由
    router := gin.New()
    userGroup := router.Group("/api/v1/users")
    userGroup.POST("/", components.UserHandler.CreateUser)
    userGroup.GET("/:id", components.UserHandler.GetUser)
    
    t.Run("创建用户成功", func(t *testing.T) {
        // 准备请求数据
        userReq := map[string]interface{}{
            "name":      "测试用户",
            "email":     "test@example.com",
            "phone":     "+86-13800138000",
            "tenant_id": "1",
        }
        
        reqBody, _ := json.Marshal(userReq)
        
        // 创建 HTTP 请求
        req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(string(reqBody)))
        req.Header.Set("Content-Type", "application/json")
        
        // 执行请求
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        // 验证响应
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var resp response.Response
        err := json.Unmarshal(w.Body.Bytes(), &resp)
        require.NoError(t, err)
        
        assert.Equal(t, 0, resp.Code)
        assert.Equal(t, "创建成功", resp.Message)
        assert.NotNil(t, resp.Data)
        
        // 验证返回的用户数据
        userData, ok := resp.Data.(map[string]interface{})
        require.True(t, ok)
        assert.Equal(t, "测试用户", userData["name"])
        assert.Equal(t, "test@example.com", userData["email"])
    })
    
    t.Run("参数验证失败", func(t *testing.T) {
        // 无效的请求数据（缺少必需字段）
        userReq := map[string]interface{}{
            "email": "invalid-email", // 无效邮箱格式
        }
        
        reqBody, _ := json.Marshal(userReq)
        req := httptest.NewRequest("POST", "/api/v1/users", strings.NewReader(string(reqBody)))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusBadRequest, w.Code)
        
        var resp response.Response
        err := json.Unmarshal(w.Body.Bytes(), &resp)
        require.NoError(t, err)
        
        assert.NotEqual(t, 0, resp.Code)
        assert.Contains(t, resp.Message, "验证失败")
    })
}
```

### 认证授权集成测试

```go
func TestAuthenticationIntegration(t *testing.T) {
    db, cleanup := test.SetupTestDB(t)
    if db == nil {
        return
    }
    defer cleanup()
    
    // 种子测试数据
    test.SeedTestData(db)
    
    testLogger, _ := test.NewTestLogger()
    components := test.NewTestComponents(db, testLogger)
    
    // 设置完整路由（包含认证中间件）
    cfg := test.NewTestConfig()
    router := routes.SetupRoutes(
        cfg, testLogger,
        components.UserHandler,
        components.CaptchaHandler,
        components.PermissionHandler,
        components.RoleHandler,
        components.FieldPermissionHandler,
        components.AuthMiddleware,
    )
    
    t.Run("完整认证流程", func(t *testing.T) {
        // 1. 获取验证码
        req := httptest.NewRequest("GET", "/api/v1/captcha/generate", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var captchaResp response.Response
        err := json.Unmarshal(w.Body.Bytes(), &captchaResp)
        require.NoError(t, err)
        
        captchaData := captchaResp.Data.(map[string]interface{})
        captchaID := captchaData["captcha_id"].(string)
        
        // 2. 登录获取 Token
        loginReq := map[string]interface{}{
            "email":          "admin@system.test",
            "password":       "admin123",
            "captcha_id":     captchaID,
            "captcha_answer": "test", // 测试环境固定答案
        }
        
        loginBody, _ := json.Marshal(loginReq)
        req = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(loginBody)))
        req.Header.Set("Content-Type", "application/json")
        
        w = httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var loginResp response.Response
        err = json.Unmarshal(w.Body.Bytes(), &loginResp)
        require.NoError(t, err)
        
        loginData := loginResp.Data.(map[string]interface{})
        accessToken := loginData["access_token"].(string)
        
        // 3. 使用 Token 访问受保护资源
        req = httptest.NewRequest("GET", "/api/v1/users/profile", nil)
        req.Header.Set("Authorization", "Bearer "+accessToken)
        
        w = httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var profileResp response.Response
        err = json.Unmarshal(w.Body.Bytes(), &profileResp)
        require.NoError(t, err)
        
        assert.Equal(t, 0, profileResp.Code)
        assert.NotNil(t, profileResp.Data)
    })
    
    t.Run("无效 Token 访问被拒绝", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v1/users/profile", nil)
        req.Header.Set("Authorization", "Bearer invalid-token")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
    })
}
```

## 🎭 Mock 和 Stub

### 生成 Mock

```bash
# 安装 mockery
go install github.com/vektra/mockery/v2@latest

# 生成所有接口的 Mock
mockery --all --output mocks --case underscore
```

### 使用 Mock 进行测试

```go
// 生成的 Mock 位于 mocks/ 目录
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func TestUserService_WithMock(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := NewUserService(mockRepo, logger, txManager)
    
    // 设置 Mock 期望
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
        Return(nil).
        Run(func(args mock.Arguments) {
            user := args.Get(1).(*models.User)
            user.ID = "generated-id" // 模拟数据库生成的ID
        })
    
    user := &models.User{Name: "Test User"}
    err := service.CreateUser(context.Background(), user)
    
    assert.NoError(t, err)
    assert.Equal(t, "generated-id", user.ID)
    mockRepo.AssertExpectations(t)
}
```

## 🗄️ 数据库测试

### 测试数据库设置

```go
// test/test_helpers.go 中的数据库设置
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
    cfg := NewTestConfig()
    testLogger, _ := NewTestLogger()
    
    // 连接测试数据库
    db, err := database.NewMySQLConnection(cfg.Database, testLogger.Logger)
    if err != nil {
        t.Skipf("跳过测试，数据库连接失败: %v", err)
        return nil, nil
    }
    
    // 自动迁移
    if err := database.AutoMigrate(db); err != nil {
        t.Skipf("跳过测试，数据库迁移失败: %v", err)
        return nil, nil
    }
    
    // 清理函数
    cleanup := func() {
        CleanupTestData(db)
        sqlDB, _ := db.DB()
        if sqlDB != nil {
            sqlDB.Close()
        }
    }
    
    return db, cleanup
}
```

### 事务测试模式

```go
func TestWithTransaction(t *testing.T) {
    db, cleanup := test.SetupTestDB(t)
    if db == nil {
        return
    }
    defer cleanup()
    
    // 在事务中运行测试
    err := db.Transaction(func(tx *gorm.DB) error {
        // 使用事务进行测试
        repo := repositories.NewUserRepository(tx, txManager, logger)
        
        user := &models.User{
            Name:  "事务测试用户",
            Email: "tx@example.com",
        }
        
        err := repo.Create(context.Background(), user)
        assert.NoError(t, err)
        
        // 验证数据在事务中存在
        var count int64
        tx.Model(&models.User{}).Where("email = ?", user.Email).Count(&count)
        assert.Equal(t, int64(1), count)
        
        // 返回错误让事务回滚（清理测试数据）
        return errors.New("rollback for test")
    })
    
    // 验证事务已回滚
    assert.Error(t, err)
    var count int64
    db.Model(&models.User{}).Where("email = ?", "tx@example.com").Count(&count)
    assert.Equal(t, int64(0), count)
}
```

## 📊 测试覆盖率

### 生成覆盖率报告

```bash
# 运行测试并生成覆盖率
go test -v -cover ./...

# 生成详细覆盖率报告
go test -v -coverprofile=coverage.out ./...

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out -o coverage.html

# 查看覆盖率统计
go tool cover -func=coverage.out
```

### 覆盖率要求

| 层级 | 最低覆盖率 | 目标覆盖率 |
|------|-----------|-----------|
| Handler | 60% | 80% |
| Service | 80% | 90% |
| Repository | 70% | 85% |
| 核心业务逻辑 | 90% | 100% |
| 整体项目 | 70% | 85% |

## 🚀 测试工作流

### 开发中的测试流程

```bash
# 1. 运行快速单元测试
go test -short ./...

# 2. 运行特定包的测试
go test -v ./internal/services/

# 3. 运行特定测试
go test -v -run TestUserService_CreateUser ./internal/services/

# 4. 运行集成测试
go test -v ./test/

# 5. 完整测试套件
make test

# 6. 质量检查（包含测试）
bash scripts/quality-check.sh
```

### CI/CD 测试管道

```yaml
# .github/workflows/test.yml 示例
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: testpass
          MYSQL_DATABASE: shield_test
        ports:
          - 3306:3306
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## 🐛 测试调试

### 调试失败的测试

```bash
# 详细输出
go test -v -run TestFailingTest

# 显示完整日志
go test -v -args -logtostderr=true

# 竞态条件检测
go test -race ./...

# 保留测试二进制文件用于调试
go test -c
./package.test -test.v -test.run TestSpecificTest
```

### 常见问题处理

**数据库连接问题**:
```go
func TestWithDBCheck(t *testing.T) {
    db, cleanup := test.SetupTestDB(t)
    if db == nil {
        t.Skip("数据库不可用，跳过测试")
        return
    }
    defer cleanup()
    
    // 测试逻辑
}
```

**时间相关测试**:
```go
func TestTimeDependent(t *testing.T) {
    // 使用固定时间进行测试
    fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
    
    // Mock 时间函数或使用依赖注入
    timeProvider := &MockTimeProvider{
        NowFunc: func() time.Time { return fixedTime },
    }
    
    service := NewServiceWithTimeProvider(timeProvider)
    // 测试逻辑
}
```

## 📋 测试检查清单

### 代码提交前检查

- [ ] 所有新代码都有对应的单元测试
- [ ] 测试覆盖率达到要求（70%+）
- [ ] 所有测试都能通过
- [ ] 没有竞态条件（`go test -race`）
- [ ] 集成测试验证了关键业务流程
- [ ] Mock 使用正确，没有过度Mock
- [ ] 测试数据能正确清理
- [ ] 测试名称清晰表达了测试意图

### 性能测试检查

- [ ] 关键路径有性能基准测试
- [ ] 数据库查询性能测试
- [ ] 并发场景测试
- [ ] 内存泄漏检测
- [ ] 响应时间基准

## 📚 测试资源

### 推荐工具

- **testify**: 断言和Mock框架
- **gomock**: 接口Mock生成
- **httptest**: HTTP测试工具
- **go-sqlmock**: SQL Mock工具
- **gofakeit**: 测试数据生成
- **testcontainers**: 容器化集成测试

### 学习资源

- [Go 测试官方文档](https://golang.org/pkg/testing/)
- [Testify 使用指南](https://github.com/stretchr/testify)
- [Go 测试最佳实践](https://github.com/golang/go/wiki/TestComments)

## 📖 相关文档

- 🏗️ [架构设计规范](./architecture.md) - 了解可测试的架构设计
- 🔧 [API 开发指南](./api-guide.md) - API 测试策略和方法
- 🚀 [快速开始指南](./getting-started.md) - 项目环境搭建

---

**重要提醒**：测试是保证代码质量的重要手段，不是可有可无的。每一行核心业务代码都应该有对应的测试，确保系统的可靠性和可维护性。