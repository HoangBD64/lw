# Copilot Instructions for LegalSwift Backend

## Architecture Overview

This is a Go backend using **Fiber framework** with a **simple repository pattern** (not clean architecture with service layer). The codebase follows these key principles:

- **Direct repository access from handlers** - no service layer abstraction
- **GORM for database operations** with PostgreSQL
- **SSO integration** using Casdoor for authentication
- **File storage** with MinIO S3-compatible storage
- **Email notifications** with cron-based scheduling

## Project Structure

```
cmd/server/main.go          # Entry point - calls app.Setup() then router.Setup()
app/                        # Core application setup
├── config.go              # YAML config loading, SSO client setup
├── database.go            # GORM setup with custom slog logger
└── logging.go             # Structured logging configuration
internal/
├── domain_model/model/     # GORM models (Admin, User, Transaction, etc.)
├── handler/               # Fiber HTTP handlers - business logic here
├── middleware/            # Authentication, CORS, logging
└── repo/                  # Direct GORM database access functions
router/router.go           # Fiber app setup, middleware, route definitions
```

## Key Patterns

### Repository Pattern
Direct repository functions without interfaces:
```go
// In repo/user.go
func GetUserByID(id uuid.UUID) (model.User, error) {
    user := model.User{}
    tx := app.Database.DB.Where("id = ?", id).First(&user)
    return user, tx.Error
}
```

### Handler Pattern
Handlers contain business logic and call repos directly:
```go
func AdminGetUser(c *fiber.Ctx) error {
    var query repo.Query
    query.Parse(c)  // Parses pagination/filters from query params
    
    users, total, err := repo.AdminGetUser(query)
    if err != nil {
        return c.JSON(base.BaseResponse{Status: false, Message: "Can't GetUser", Error: err})
    }
    
    return c.JSON(base.BaseResponse{Status: true, Data: fiber.Map{"users": users, "total": total}})
}
```

### Query Pattern
Standardized pagination and filtering:
```go
type Query struct {
    UserID  *uuid.UUID  `query:"user_id"`
    Limit   int         `query:"limit"`
    Page    int         `query:"page"`
    Search  string      `query:"search"`
    // ... many more filter fields
}

func (query *Query) Parse(c *fiber.Ctx) {
    // Auto-parses query parameters with defaults
}
```

### Model Embedding
All models embed base `Model` struct:
```go
type User struct {
    Model    `gorm:"embedded"`  // ID, CreatedAt, UpdatedAt
    Name     string
    Email    *string  `gorm:"Index"`
    SSO_ID   *string  `gorm:"Index"`
}
```

## Development Workflows

### Local Development
```bash
# Start with air for hot reloading
air

# Database setup via docker-compose
docker-compose up -d  # PostgreSQL container
```

### Configuration
- `config.yaml` for app settings
- `.env` files for secrets (DB_HOST, DB_PASSWORD, SSO credentials)
- Environment variables override YAML config

### Testing
- Direct repo testing against database
- No mocking - test with real DB connections
- Focus on integration testing over unit tests

## SSO Integration (Casdoor)
- User authentication via Casdoor SSO
- Admin authentication with role-based permissions
- JWT token validation in middleware
- Permission system with View/Manage granularity per module

## File Storage (MinIO)
- S3-compatible API for document uploads
- Configured via MinIO client in app setup
- Direct integration in handlers for file operations

## Code Conventions

### Error Handling
```go
// Standard error response pattern
return c.JSON(base.BaseResponse{
    Status: false, 
    Message: "Description", 
    Error: err
})
```

### Pagination
Always use `repo.Query` struct for list endpoints with built-in pagination.

### Transactions
Use GORM transactions for multi-table operations, wrap in repository functions.

### Validation
GORM model tags for database constraints, manual validation in handlers.

## Dependencies
- **Fiber v2** - HTTP framework
- **GORM** - ORM with PostgreSQL driver
- **MinIO Go SDK** - S3-compatible storage
- **Casdoor SDK** - SSO integration
- **Robfig Cron** - Scheduled jobs
- **Go-Mail** - Email sending

## Related Projects
This backend pairs with a Nuxt 4 frontend using similar auth patterns and API structure.