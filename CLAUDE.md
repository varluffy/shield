# CLAUDE.md

This file provides specialized guidance to Claude Code (claude.ai/code) when working with code in this repository, focusing on AI-assisted development workflows and patterns.

## 🚨 CRITICAL DATABASE RULE 🚨
**NEVER USE DIRECT MySQL COMMANDS - ONLY USE MCP TOOL dbhub-mysql-npx **
- ❌ `mysql -u root -p shield` ← FORBIDDEN
- ❌ `mysql < file.sql` ← FORBIDDEN  
- ❌ `mysqldump shield` ← FORBIDDEN
- ✅ Use MCP tool interface for ALL database operations

## 📖 Quick Reference

**Project Type**: Go microservices framework with clean architecture  
**Key Technologies**: Gin, Wire, GORM, MySQL, Redis, OpenTelemetry  
**Documentation**: See [docs/development/](docs/development/) for comprehensive guides  
**字段权限系统**: See [docs/development/field-permissions-guide.md](docs/development/field-permissions-guide.md) for field permission system

## 🤖 AI Development Patterns

### Essential Command Reference
These commands are optimized for AI-assisted development workflows:

```bash
# 🚀 Quick Start (AI-Assisted Development)
make quick-init        # Fast setup for AI development session
make run               # Start development server (auto-cleanup)
make wire              # Regenerate DI code after changes
make wire && make test  # Full validation

# 🔧 AI Development Iteration
make wire && make test # Essential after adding new components
make status           # Check if services are running
make stop-service     # Clean stop when switching contexts

# 🧪 Testing Patterns (AI-Optimized)
go test -v ./test/ -run TestCaptcha           # Test specific features
go test -v ./test/ -run TestPermission        # Permission system tests
go test -v ./test/ -run TestFieldPermission   # Field permission tests
go test -v -cover ./internal/services/        # Service layer coverage
make wire && make test                # Comprehensive validation
```

### AI Development Workflow
```bash
# Typical AI-assisted development cycle:
1. make wire          # After adding new constructors/providers
2. make test          # Validate changes
3. make run           # Test running system
4. make docs          # Update API documentation
```

### Database Operations (AI-Optimized)
**🚨 CRITICAL: ALWAYS Use MCP Tool for MySQL Operations 🚨**

**❌ NEVER USE THESE COMMANDS:**
```bash
# ❌ FORBIDDEN - DO NOT USE
mysql -u root -p shield
mysql -h localhost -u root -p -e "SELECT * FROM users;"
mysqldump shield > backup.sql

# ❌ THESE ARE ALSO FORBIDDEN  
./scripts/mysql_query.sh
sudo mysql shield
```

**✅ ALWAYS USE MCP TOOL:**
```sql
-- ✅ CORRECT WAY - Use MCP tool for ALL database queries
SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'shield';
SELECT tenant_id, COUNT(*) as user_count FROM users GROUP BY tenant_id;
SELECT * FROM permissions WHERE scope = 'tenant' LIMIT 10;
SHOW DATABASES LIKE 'shield';
```

**🎯 Why MCP Tool Only:**
- ✅ Proper connection management and security
- ✅ Consistent authentication and permissions  
- ✅ Integrated with Claude Code environment
- ✅ Prevents accidental data corruption
- ❌ Direct MySQL commands bypass security controls

## 🎯 AI-Focused Architecture Patterns

### Critical Rules for AI Development
- **🚨 DATABASE OPERATIONS**: NEVER use direct MySQL commands - ONLY use MCP tool
- **Wire Dependency Injection**: ALWAYS run `make wire` after modifying providers
- **Clean Architecture**: Handler → Service → Repository (never skip layers)
- **Multi-Tenant Context**: All operations must include tenant isolation
- **Context Propagation**: Pass `context.Context` through all layers for tracing

### AI Development Anti-Patterns to Avoid

**🚨 NEVER Use Direct MySQL Commands:**
```bash
# ❌ ABSOLUTELY FORBIDDEN - These will cause issues
mysql -u root -p shield -e "INSERT INTO users..."
mysql shield < migration.sql
./any_mysql_script.sh

# ✅ CORRECT - Use MCP tool only
# Use MCP tool interface for ALL database operations
```

**❌ Architecture Violations:**
```go
// ❌ DON'T: Handler calling Repository directly
func (h *UserHandler) GetUser(c *gin.Context) {
    user := h.userRepo.Get(id) // VIOLATION!
}

// ✅ DO: Handler → Service → Repository
func (h *UserHandler) GetUser(c *gin.Context) {
    user, err := h.userService.GetUser(c.Request.Context(), id)
}
```

### Essential Configuration (AI Development)
```yaml
# Minimal config for AI development sessions:
database:
  name: "shield"
  user: "root"
  password: "123456"  # Update for your environment

# Optional services (auto-fallback):
redis:
  addrs: ["localhost:6379"]  # Validation: captcha storage
```

## 🏗️ AI Development Workflow

### New Feature Development (AI-Optimized Pattern)
```bash
# 1. Create Model (internal/models/)
# 2. Create Repository Interface + Implementation (internal/repositories/)
# 3. Add to RepositoryProviderSet (internal/repositories/providers.go)
# 4. Create Service Interface + Implementation (internal/services/)
# 5. Add to ServicesProviderSet (internal/services/providers.go)
# 6. Create Handler (internal/handlers/)
# 7. Add to HandlersProviderSet (internal/handlers/providers.go)
# 8. Register Route (internal/routes/routes.go)
# 9. CRITICAL: make wire  # Regenerate dependency injection
# 10. make test          # Validate implementation
```

### AI Code Generation Patterns
```go
// When adding new entity, follow this exact pattern:

// 1. Model (internal/models/entity.go)
type Entity struct {
    ID        string    `gorm:"primarykey" json:"id"`
    TenantID  string    `gorm:"not null;index" json:"tenant_id"`  // REQUIRED
    Name      string    `gorm:"not null" json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 2. Repository Interface (internal/repositories/entity.go)
type EntityRepository interface {
    Create(ctx context.Context, entity *models.Entity) error
    GetByID(ctx context.Context, id string) (*models.Entity, error)
    GetByTenant(ctx context.Context, tenantID string) ([]*models.Entity, error)
}

// 3. Service Interface (internal/services/entity.go)
type EntityService interface {
    CreateEntity(ctx context.Context, entity *models.Entity) error
    GetEntity(ctx context.Context, id string) (*models.Entity, error)
}

// 4. Handler (internal/handlers/entity.go)
type EntityHandler struct {
    entityService services.EntityService
    logger        logger.Logger
}
```

## 🐛 AI Debugging & Troubleshooting

### Common AI Development Issues

**1. Wire Generation Failures**
```bash
# Symptom: Build fails with missing providers
# Solution: 
make wire              # Regenerate dependency injection
# Check: All new constructors added to correct ProviderSet
```

**2. Port Conflicts During Development**
```bash
# Symptom: "port already in use" errors
# Solution:
make stop-service      # Clean shutdown
make kill-port         # Force kill port 8080
make status           # Verify clean state
```

**3. Database Connection Issues**
```bash
# Quick diagnostics:
# 1. Check MySQL is running: systemctl status mysql
# 2. ✅ Test connection via MCP tool ONLY:
SELECT 1;  # Should return 1 if connected

# 3. ✅ Verify database exists via MCP tool:
SHOW DATABASES LIKE 'shield';

# ❌ NEVER use these commands:
# mysql -u root -p shield  <-- FORBIDDEN
# mysql shield -e "SELECT 1;"  <-- FORBIDDEN
```

**4. Multi-Tenant Context Missing**
```go
// Symptom: Data leaks between tenants
// Check: All models have tenant_id
type Entity struct {
    TenantID string `gorm:"not null;index" json:"tenant_id"` // REQUIRED!
}

// Check: Services use tenant filtering
func (s *service) GetEntities(ctx context.Context) ([]*Entity, error) {
    tenantID := getTenantIDFromContext(ctx) // REQUIRED!
    return s.repo.GetByTenant(ctx, tenantID)
}
```

### AI Development Gotchas

1. **🚨 DATABASE ACCESS**: NEVER use direct MySQL commands - ONLY MCP tool (this is critical!)
2. **Wire Provider Order**: Add providers to correct ProviderSet files
3. **Context Propagation**: Always pass `context.Context` through all layers
4. **Tenant Isolation**: Every user operation MUST include tenant_id filtering
5. **Testing**: Run `make test` before committing

**⚠️ Common MySQL Command Mistakes to Avoid:**
```bash
# ❌ These are all FORBIDDEN:
mysql -u root -p
mysql shield
./run_sql.sh
mysqldump shield
mysql < backup.sql

# ✅ ONLY use MCP tool for ANY database operation
```

### Quick Validation Commands
```bash
# After making changes:
make wire && make test  # Essential validation
make wire && make test  # Full quality check

# For debugging specific issues:
make status            # Check service health
go test -v ./test/ -run TestPermission  # Test specific subsystem
```

### 🔐 Environment-Aware Authentication System

**New Unified Login System** (🚨 test-login interface removed for security):
- **Development Environment**: Supports captcha bypass with `captcha_id: "dev-bypass"` and `answer: "dev-1234"`
- **Production Environment**: Always requires valid captcha verification
- **Single Login Endpoint**: `/api/v1/auth/login` handles all environments intelligently

**Security Configuration**:
```yaml
# Development (configs/config.dev.yaml)
auth:
  captcha_mode: "flexible"    # Allows bypass
  dev_bypass_code: "dev-1234" # Bypass answer

# Production (configs/config.prod.yaml) 
auth:
  captcha_mode: "strict"      # Enforces captcha
  dev_bypass_code: ""         # No bypass available
```

## 🧪 AI Testing Patterns

### Standard Test Users (Essential for AI Development)
**🎯 Use these pre-configured test users to avoid login debugging:**

```bash
# Create standard test users with known passwords
go run cmd/migrate/*.go -action=create-test-users -config=configs/config.dev.yaml

# List test user status
go run cmd/migrate/*.go -action=list-test-users -config=configs/config.dev.yaml
```

**Available Test Users:**
- **System Admin**: `admin@system.test` / `admin123` (tenant_id: 0, bypasses all permissions)
- **Tenant Admin**: `admin@tenant.test` / `admin123` (tenant_id: 1)
- **Test User**: `test@example.com` / `test123` (tenant_id: 1)
- **Regular User**: `user@tenant.test` / `user123` (tenant_id: 1)

### AI Development Testing Pattern
```bash
# 1. Get test token (uses development captcha bypass) - SYSTEM ADMIN (recommended for development)
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"admin@system.test",
    "password":"admin123",
    "tenant_id":"0",
    "captcha_id":"dev-bypass",
    "answer":"dev-1234"
  }' | jq -r '.data.access_token')

# 2. Test with authorization  
curl -H "Authorization: Bearer $JWT_TOKEN" "http://localhost:8080/api/v1/users/profile"

# 3. For tenant-specific testing
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"test@example.com",
    "password":"test123",
    "tenant_id":"1",
    "captcha_id":"dev-bypass",
    "answer":"dev-1234"
  }' | jq -r '.data.access_token')
```

### Service Layer Unit Testing (AI-Optimized)
**🏆 Shield has comprehensive service layer unit tests covering 4 core services:**

```bash
# Run all service unit tests (1400+ lines of test code)
go test -v ./test/ -run ".*ServiceUnitTests"

# Run specific service tests
go test -v ./test/ -run TestUserServiceUnitTests        # User service tests
go test -v ./test/ -run TestPermissionServiceUnitTests  # Permission service tests  
go test -v ./test/ -run TestRoleServiceUnitTests        # Role service tests
go test -v ./test/ -run TestBlacklistServiceUnitTests   # Blacklist service tests

# Run specific test scenarios
go test -v ./test/ -run "TestUserService.*CreateUser"       # All CreateUser tests
go test -v ./test/ -run "TestPermissionService.*IsSystem"   # System admin tests
go test -v ./test/ -run "TestRoleService.*AssignPermissions" # Permission assignment tests
```

**Service Testing Architecture:**
- ✅ **Real Database Connections**: No mocks, ensures integration reliability
- ✅ **Standard Test Users**: Consistent authentication across all tests
- ✅ **Comprehensive Error Scenarios**: Input validation, business logic, resource not found
- ✅ **Full CRUD Coverage**: Create, Read, Update, Delete operations tested
- ✅ **Tenant Isolation Testing**: Multi-tenant context validation

### Service Testing Helpers (AI Code Generation)
**🔧 Use these test helper patterns when adding new services:**

```go
// 1. Setup test database and standard users
db, cleanup := SetupTestDB(t)
if db == nil {
    return // Skip if database unavailable
}
defer cleanup()

// 2. Get standard test users (avoid creating new users)
testUsers := SetupStandardTestUsers(db)
systemAdmin := testUsers["admin@system.test"]
tenantAdmin := testUsers["admin@tenant.test"]
regularUser := testUsers["user@tenant.test"]

// 3. Create test components (full service stack)
testLogger, err := NewTestLogger()
require.NoError(t, err)
components := NewTestComponents(db, testLogger)

// 4. Generate JWT for authentication testing
token, err := GenerateTestJWT(components, systemAdmin.UUID, "0")
require.NoError(t, err)
authHeaders := CreateAuthHeader(token)
```

**Essential Testing Pattern for New Services:**
```go
func TestNewServiceUnitTests(t *testing.T) {
    // Standard setup (copy from existing service tests)
    db, cleanup := SetupTestDB(t)
    if db == nil { return }
    defer cleanup()
    
    testUsers := SetupStandardTestUsers(db)
    testLogger, err := NewTestLogger()
    require.NoError(t, err)
    components := NewTestComponents(db, testLogger)
    
    // Test success scenarios
    t.Run("Test Create Success", func(t *testing.T) {
        // Test successful creation with valid data
    })
    
    // Test error scenarios  
    t.Run("Test Create Invalid Input", func(t *testing.T) {
        // Test validation errors, missing fields, etc.
    })
    
    // Test business logic
    t.Run("Test Business Logic", func(t *testing.T) {
        // Test specific business rules and constraints
    })
}
```

### AI Testing Scripts
- `make test` - Run all tests (use before commits)
- `make wire && make test` - Full validation cycle (essential after code changes)
- `scripts/test_permissions.sh` - Permission system validation
- `go test -v ./test/ -run TestCaptcha` - Feature-specific testing
- `go test -v ./test/ -run TestFieldPermissionServiceUnitTests` - Field permission unit tests

## 📚 Documentation References

For comprehensive guidance, see the structured documentation:

- **[Getting Started](docs/development/getting-started.md)** - Project setup and quick start
- **[Architecture Guide](docs/development/architecture.md)** - Detailed architecture rules and patterns  
- **[API Development](docs/development/api-guide.md)** - API design and implementation standards
- **[Field Permissions Guide](docs/development/field-permissions-guide.md)** - Field-level permission system guide 🛡️
- **[Testing Cheatsheet](docs/development/testing-cheatsheet.md)** - Quick reference for testing commands ⚡
- **[Testing Guide](docs/development/testing-guide.md)** - Testing strategies and best practices
- **[Service Testing Guide](docs/development/service-testing-guide.md)** - Service layer unit testing patterns ✨
- **[Test Users Guide](docs/development/test-users.md)** - Standard test users and authentication

## 🚨 Critical AI Reminders

1. **🚨🚨 NEVER USE DIRECT MySQL COMMANDS 🚨🚨** - ONLY use MCP tool for database operations
2. **ALWAYS run `make wire`** after modifying any constructor or provider
3. **Include tenant_id** in all user data models and queries
4. **Pass context.Context** through all function calls for tracing
5. **Use standard test users** for authentication testing: `admin@system.test` / `admin123`
6. **Follow service testing patterns** when adding new services (see service-testing-guide.md)
7. **Run comprehensive tests** before suggesting code changes: `make wire && make test`
8. **Use real database connections** in service tests (no mocks) for integration reliability

**💀 ABSOLUTE PROHIBITION - DO NOT USE:**
```bash
mysql -u root -p shield  # ❌ FORBIDDEN
mysql < file.sql         # ❌ FORBIDDEN  
mysqldump shield         # ❌ FORBIDDEN
./mysql_scripts.sh       # ❌ FORBIDDEN
```

**✅ ONLY ALLOWED DATABASE ACCESS:**
- Use MCP tool interface ONLY
- All SQL queries through MCP tool
- No exceptions to this rule

---

**AI Development Note**: This project follows strict clean architecture. When adding new features, always maintain the Handler → Service → Repository flow and ensure proper dependency injection via Wire.