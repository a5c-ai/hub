# Hub Backend

This is the Go backend for the Hub git hosting service.

## Project Structure

```
/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── api/             # HTTP API routes and handlers
│   ├── auth/            # Authentication and JWT handling
│   ├── config/          # Configuration management
│   ├── db/              # Database connection and setup
│   ├── middleware/      # HTTP middleware
│   └── models/          # Database models
├── pkg/                 # Public packages (empty for now)
├── migrations/          # Database migrations (empty for now)
├── config.example.yaml  # Example configuration file
├── Dockerfile           # Multi-stage Docker build
└── go.mod              # Go module definition
```

## Features

- ✅ HTTP server with Gin framework
- ✅ PostgreSQL database integration with GORM
- ✅ JWT authentication foundation
- ✅ Environment-based configuration with Viper
- ✅ Structured logging with Logrus
- ✅ CORS middleware
- ✅ Health check endpoint
- ✅ Graceful shutdown
- ✅ Docker support
- ✅ Basic unit tests

## Quick Start

1. **Setup development environment:**
   ```bash
   ./scripts/dev-setup.sh
   ```

2. **Configure the application:**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

3. **Start PostgreSQL** (required):
   ```bash
   # Using Docker
   docker run --name hub-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_USER=hub -e POSTGRES_DB=hub -p 5432:5432 -d postgres:15
   ```

4. **Run the application:**
   ```bash
   go run ./cmd/server
   ```

## API Endpoints

### Health Check
- `GET /health` - System health status

### API v1
- `GET /api/v1/ping` - Simple ping endpoint
- `POST /api/v1/auth/login` - User login (placeholder)
- `POST /api/v1/auth/register` - User registration (placeholder)
- `POST /api/v1/auth/logout` - User logout (placeholder)

### Protected Endpoints (require JWT token)
- `GET /api/v1/profile` - Get user profile
- `GET /api/v1/repositories/` - List repositories (placeholder)
- `POST /api/v1/repositories/` - Create repository (placeholder)
- `GET /api/v1/repositories/:owner/:repo` - Get repository (placeholder)

### Admin Endpoints (require admin JWT token)
- `GET /api/v1/admin/users` - List users (placeholder)

## Configuration

Configuration can be provided via:
1. `config.yaml` file
2. Environment variables

### Environment Variables

- `ENVIRONMENT` - Environment (development/production)
- `LOG_LEVEL` - Log level (0-6, default: 4)
- `PORT` - Server port (default: 8080)
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `DB_SSLMODE` - Database SSL mode
- `JWT_SECRET` - JWT secret key
- `JWT_EXPIRATION_HOUR` - JWT expiration in hours

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o hub ./cmd/server
```

### Docker Build
```bash
docker build -t hub-backend .
```

### Docker Run
```bash
docker run -p 8080:8080 -e DB_HOST=host.docker.internal hub-backend
```

## Next Steps

This foundation provides the basic structure for the Hub backend. Next phases will include:

1. **Complete Authentication**: User registration, login, password hashing
2. **Repository Management**: Git operations, repository CRUD
3. **User Management**: User profiles, organizations, teams
4. **Pull Requests**: PR workflow, reviews, comments
5. **CI/CD Integration**: Pipeline management, runners
6. **Advanced Features**: Search, webhooks, integrations

## Architecture Notes

- **Microservices Ready**: Structure supports splitting into microservices
- **Database Agnostic**: GORM allows easy database switching
- **Cloud Native**: Designed for containerization and Kubernetes
- **Security First**: JWT authentication, prepared for OAuth2/SAML
- **Scalable**: Connection pooling, graceful shutdown, health checks