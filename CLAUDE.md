# CLAUDE.md

This file provides specialized guidance to Claude Code (claude.ai/code) when working with code in this repository, focusing on AI-assisted development workflows and patterns.

## 📖 Quick Reference

**Project Type**: Go microservices framework with clean architecture  
**Key Technologies**: Gin, Wire, GORM, MySQL, Redis, OpenTelemetry  
**Documentation**: See [docs/development/](docs/development/) for comprehensive guides

## 🤖 AI Development Patterns

### Essential Command Reference
These commands are optimized for AI-assisted development workflows:

```bash
# 🚀 Quick Start (AI-Assisted Development)
make quick-init        # Fast setup for AI development session
make run               # Start development server (auto-cleanup)
make wire              # Regenerate DI code after changes
bash scripts/quality-check.sh  # Full quality validation

# 🔧 AI Development Iteration
make wire && make test # Essential after adding new components
make status           # Check if services are running
make stop-service     # Clean stop when switching contexts

# 🧪 Testing Patterns (AI-Optimized)
go test -v ./test/ -run TestCaptcha           # Test specific features
go test -v ./test/ -run TestPermission        # Permission system tests
go test -v -cover ./internal/services/        # Service layer coverage
bash scripts/quality-check.sh                # Comprehensive validation
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
**🤖 Use MCP tool for all database queries** - Never use direct `mysql` commands
```sql
-- Common AI development queries:
SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'shield';
SELECT tenant_id, COUNT(*) as user_count FROM users GROUP BY tenant_id;
SELECT * FROM permissions WHERE scope = 'tenant' LIMIT 10;
```

## 🎯 AI-Focused Architecture Patterns

### Critical Rules for AI Development
- **Wire Dependency Injection**: ALWAYS run `make wire` after modifying providers
- **Clean Architecture**: Handler → Service → Repository (never skip layers)
- **Multi-Tenant Context**: All operations must include tenant isolation
- **Context Propagation**: Pass `context.Context` through all layers for tracing

### AI Development Anti-Patterns to Avoid
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
# 2. Test connection via MCP tool:
SELECT 1;  # Should return 1 if connected

# 3. Verify database exists:
SHOW DATABASES LIKE 'shield';
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

1. **Wire Provider Order**: Add providers to correct ProviderSet files
2. **Context Propagation**: Always pass `context.Context` through all layers
3. **Tenant Isolation**: Every user operation MUST include tenant_id filtering
4. **Database Transactions**: Use MCP tool, never direct MySQL commands
5. **Testing**: Run `bash scripts/quality-check.sh` before committing

### Quick Validation Commands
```bash
# After making changes:
make wire && make test  # Essential validation
bash scripts/quality-check.sh  # Full quality check

# For debugging specific issues:
make status            # Check service health
go test -v ./test/ -run TestPermission  # Test specific subsystem
```

## 🧪 AI Testing Patterns

### Development Testing Workflow
```bash
# Quick API testing pattern for AI development:
# 1. Get test token (bypasses captcha)
curl -X POST "http://localhost:8080/api/v1/auth/test-login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123","tenant_id":"1"}'

# 2. Extract token and test endpoints
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/test-login" \
  -d '{"email":"admin@example.com","password":"admin123","tenant_id":"1"}' | \
  jq -r '.data.access_token')

# 3. Test with authorization
curl -H "Authorization: Bearer $JWT_TOKEN" "http://localhost:8080/api/v1/users/profile"
```

### AI Testing Scripts
- `bash scripts/quality-check.sh` - Comprehensive validation (use before commits)
- `scripts/test_permissions.sh` - Permission system validation
- `go test -v ./test/ -run TestCaptcha` - Feature-specific testing

## 📚 Documentation References

For comprehensive guidance, see the structured documentation:

- **[Getting Started](docs/development/getting-started.md)** - Project setup and quick start
- **[Architecture Guide](docs/development/architecture.md)** - Detailed architecture rules and patterns  
- **[API Development](docs/development/api-guide.md)** - API design and implementation standards
- **[Testing Guide](docs/development/testing-guide.md)** - Testing strategies and best practices

## 🚨 Critical AI Reminders

1. **ALWAYS run `make wire`** after modifying any constructor or provider
2. **Use MCP tool** for all database operations (never direct mysql commands)
3. **Include tenant_id** in all user data models and queries
4. **Pass context.Context** through all function calls for tracing
5. **Run quality checks** before suggesting code changes: `bash scripts/quality-check.sh`

---

**AI Development Note**: This project follows strict clean architecture. When adding new features, always maintain the Handler → Service → Repository flow and ensure proper dependency injection via Wire.