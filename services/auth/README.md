# Auth Microservice with Chi HTTP Server

A complete authentication microservice built with Go, Chi router, and gRPC support. Includes both HTTP REST API and gRPC interfaces.

## Features

- **HTTP REST API** using Chi router
- **gRPC service** for service-to-service communication
- **JWT authentication** with access and refresh tokens
- **Redis integration** for token storage
- **PostgreSQL support** ready for database operations
- **Docker containerization**
- **Health checks** and monitoring
- **Graceful shutdown**

## API Endpoints

### Authentication

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com", 
  "password": "password123"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "roles": ["user"],
      "active": true,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  },
  "message": "Login successful"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Get User Profile (Protected)
```http
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

#### Validate Token (Protected)
```http
GET /api/v1/auth/validate
Authorization: Bearer <token>
```

### Health Check
```http
GET /health
```

## gRPC Service

The auth service also exposes a gRPC interface on port `9090` with the following methods:

- `Login(LoginRequest) returns (LoginResponse)`
- `Register(RegisterRequest) returns (RegisterResponse)`
- `ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse)`
- `RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse)`
- `GetUserProfile(GetUserProfileRequest) returns (GetUserProfileResponse)`

## Configuration

The service can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `HTTP_PORT` | HTTP server port | `8081` |
| `GRPC_PORT` | gRPC server port | `9090` |
| `JWT_SECRET` | JWT signing secret | `your-super-secret-jwt-key-change-this-in-production` |
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379` |
| `DATABASE_URL` | PostgreSQL connection URL | - |
| `JAEGER_ENDPOINT` | Jaeger tracing endpoint | - |

## Development

### Running Locally

1. **Start the auth service:**
```bash
cd services/auth
go run cmd/main.go
```

2. **With custom configuration:**
```bash
export HTTP_PORT=8081
export GRPC_PORT=9090
export JWT_SECRET=your-secret-key
export REDIS_URL=redis://localhost:6379
go run cmd/main.go
```

### Running with Docker

```bash
# Build and run with docker-compose
docker-compose up --build -d

# Check logs
docker-compose logs auth

# Stop services
docker-compose down
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   gRPC Client   │
└─────────────────┘    └─────────────────┘
         │                       │
         v                       v
┌─────────────────┐    ┌─────────────────┐
│   Chi HTTP      │    │   gRPC Server   │
│   Server :8081  │    │   :9090         │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┘
                  │
         ┌─────────────────┐
         │  Auth Service   │
         │   (Business     │
         │    Logic)       │
         └─────────────────┘
                  │
         ┌─────────────────┐
         │     Redis       │
         │  (Token Store)  │
         └─────────────────┘
```

## Security Features

- **JWT Tokens**: Short-lived access tokens (15 minutes) and long-lived refresh tokens (7 days)
- **Password Hashing**: bcrypt with salt for secure password storage
- **Token Validation**: Middleware for protecting endpoints
- **CORS Support**: Configurable CORS headers
- **Rate Limiting**: Built-in support (can be configured via middleware)

## Testing

### Manual Testing with curl/PowerShell

```powershell
# Register a new user
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/register" -Method Post -ContentType "application/json" -Body '{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}'

# Login
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"email":"test@example.com","password":"password123"}'

# Use token for protected endpoints
$token = $loginResponse.data.token
$headers = @{"Authorization" = "Bearer $token"}
Invoke-RestMethod -Uri "http://localhost:8081/api/v1/auth/profile" -Method Get -Headers $headers
```

## Default Test User

The service comes with a pre-configured test user for development:

- **Email**: `test@example.com`
- **Password**: `password123`

You can use this user to test the authentication flow immediately after starting the service.

## Deployment Notes

- The service supports graceful shutdown with proper cleanup
- Health checks are available for container orchestration
- All ports are configurable via environment variables
- Supports running behind reverse proxies like Nginx or Envoy
- Ready for horizontal scaling (stateless design with Redis for session storage)
