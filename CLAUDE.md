# CLAUDE.md

This file provides essential project context and guidance specifically for Claude Code (claude.ai/code) when working with code in this repository. It contains key information about the architecture, commands, and workflow that Claude Code needs to understand in order to effectively assist with development tasks.

## Project Overview

UltraFit is a Go-based microservices development framework implementing clean architecture with multi-tenant permission management. The project uses modern Go patterns including dependency injection (Wire), structured logging (Zap), and distributed tracing (OpenTelemetry).

## Core Architecture

### Clean Architecture Layers
The project strictly follows clean architecture with these layers:

1. **Handler Layer** (`internal/handlers/`): HTTP request handling, parameter binding
2. **Service Layer** (`internal/services/`): Business logic, transaction management  
3. **Repository Layer** (`internal/repositories/`): Data access abstractions
4. **Model Layer** (`internal/models/`): Domain entities and database models

### Key Architecture Rules
- Handlers ONLY call Services, never Repositories directly
- Services coordinate business logic and call Repositories
- All cross-layer communication uses interfaces
- Dependency injection handled by Google Wire
- Context propagation for tracing and tenant isolation

### Wire Dependency Injection
Modular provider organization with separate ProviderSet files for each layer:
- `internal/infrastructure/providers.go` - Config, DB, logging, tracing
- `internal/repositories/providers.go` - Data access layer
- `internal/services/providers.go` - Business logic layer  
- `internal/handlers/providers.go` - HTTP handlers
- `internal/middleware/providers.go` - HTTP middleware

## Essential Commands

### Development Workflow
```bash
# Quick start - run complete initialization
make init              # Project initialization (setup + migrate)
make quick-init        # Fast initialization for development
make full-setup        # Complete setup (includes admin creation)

# Service management
make run               # Start development server (auto-prep + stop old)
make safe-run          # Start with port checking (won't kill existing)
make start-service     # Background start
make stop-service      # Stop all ultrafit services
make restart-service   # Restart services
make status           # Check service status
make check-port       # Check if port 8080 is occupied
make kill-port        # Kill processes using port 8080

# Code generation and tools
make wire             # Generate dependency injection code
make docs             # Generate API documentation (swagger)
make migrate          # Run database migrations
make admin            # Run admin tools
make create-admin     # Create admin user

# Quality assurance
make test             # Run all tests
bash scripts/quality-check.sh  # Comprehensive quality check
make format           # Format code (gofmt, goimports) - **MISSING IN MAKEFILE**
make lint             # Run code linters (golangci-lint) - **MISSING IN MAKEFILE**
make full-check       # Complete quality check (test + lint) - **MISSING IN MAKEFILE**

# Build and cleanup
make build           # Build all binaries
make build-prod      # Production build (Linux)
make clean           # Clean build files
make deps            # Install dependencies
make setup           # Basic project setup

# Testing specific modules
go test -v ./test/ -run TestCaptcha           # Captcha tests
go test -v ./test/ -run TestSimplifiedAPI     # API tests
go test -v ./pkg/captcha/...                  # Package tests
go test -v -cover ./...                       # All tests with coverage
```

### Database Operations
**Important: Use MCP tool instead of direct SQL commands**
```bash
# Safe database operations via MCP
echo "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'ultrafit_dev'" | 
    mcp-execute-sql
```

## Multi-Tenant Features

### Tenant Isolation
- All user data includes `tenant_id` for isolation
- JWT tokens contain tenant context
- Database queries automatically filter by tenant
- Permission system respects tenant boundaries

### Key Multi-Tenant Models
Available via MCP tool: Users, Roles, Permissions, Field Permissions - all designed with tenant isolation

## Configuration

### Environment Setup
- Development: `configs/config.dev.yaml`
- Local override: `configs/config.local.yaml` (gitignored)
- Production: `configs/config.prod.yaml`

### Required Configuration
```yaml
# Database (required)
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_password"
  name: "ultrafit_dev"

# Optional but recommended
redis:
  addrs: ["localhost:6379"]
  
auth:
  jwt:
    secret: "your-secret-key"
```

## Code Structure Overview

```
├── cmd/                    # Application entry points
│   ├── server/            # Main web server
│   ├── migrate/           # Database migration tool
│   └── admin/             # Admin utilities
├── internal/              # Private application code
│   ├── handlers/          # HTTP request handlers (clean architecture)
│   ├── services/          # Business logic services
│   ├── repositories/      # Data access layer
│   ├── models/           # Database models
│   ├── dto/              # Data transfer objects
│   ├── middleware/       # HTTP middleware
│   ├── config/           # Configuration management
│   └── wire/             # Dependency injection
├── pkg/                  # Reusable packages
│   ├── auth/             # JWT authentication
│   ├── captcha/          # Captcha service
│   ├── logger/           # Structured logging
│   ├── response/         # HTTP response formatting
│   └── errors/           # Error handling
├── test/                 # Integration tests
├── docs/                 # Project documentation
└── configs/              # Configuration files
```

## Development Workflow

### Adding New Components
1. **New Entity**: Create model in `internal/models/` → repository in `internal/repositories/` → service in `internal/services/` → handlers in `internal/handlers/`
2. **New Endpoint**: Add service method → handler method → route in `internal/routes/routes.go`
3. **Post-Changes**: Always run `make wire` after adding constructor functions to provider sets

### Common Modification Patterns
```bash
# After adding new components
make wire    # Regenerate DI code
make test    # Run tests
bash scripts/quality-check.sh  # Full quality check (if make lint unavailable)

# Debug tenant issues safely using MCP tool
echo "SELECT tenant_id, COUNT(*) FROM users GROUP BY tenant_id LIMIT 10;" | mcp__dbhub-mysql-npx__execute_sql
```

### Quality Control
The project includes a comprehensive quality check script at `scripts/quality-check.sh` that verifies:
- Go version compatibility
- Code formatting
- Compilation success
- Wire code generation
- Dependencies
- Core tests
- Lint checks (if golangci-lint available)

### Key Development Rules from .cursorrules
- Strict clean architecture: Handler → Service → Repository
- All cross-layer communication via interfaces
- Wire dependency injection for all components  
- Context propagation for tracing and tenant isolation
- Repository pattern for data access
- Comprehensive error handling and logging

### Security Features
- JWT authentication with tenant context
- Layered permission system (menu, button, API, field level)
- Graphical captcha with Redis/memory fallback
- Automatic SQL injection prevention via ORM
- Context-aware error handling

### API Design
- Standard response format: `{"code":0,"message":"success","data":{},"timestamp":"..."}`
- RESTful endpoints with proper HTTP status codes
- Comprehensive validation using gin-validator
- OpenAPI/Swagger documentation via make docs