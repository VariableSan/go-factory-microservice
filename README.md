# Go-Factory Microservice

Monorepo workspace for scalable Go microservices.  
This snapshot includes:
- gateway — HTTP entrypoint & reverse proxy
- auth — gRPC authentication service

## Project Structure

```
├── docker-compose.yml          # Container orchestration
├── go.work                     # Go workspace definition
├── Taskfile.yml               # Task automation
├── pkg/
│   ├── common/               # Shared utilities
│   └── proto/                # Protocol buffer definitions
└── services/
    ├── auth/                 # Authentication service
    └── gateway/              # API Gateway service
```

## Getting Started

### Prerequisites

- Go 1.25.0 or later
- Docker & Docker Compose
- [Task](https://taskfile.dev) (task runner)
- Protocol Buffers compiler (protoc)

### Quick Start

1. **Setup the project:**
   ```bash
   task setup
   ```

2. **Start services:**
   ```bash
   task docker-up
   ```

3. **View available tasks:**
   ```bash
   task --list
   ```

## Available Tasks

### Development

- `task setup` - Initialize workspace and install tools
- `task install-tools` - Install development tools
- `task mod-tidy` - Run go mod tidy on all modules
- `task format` - Format all Go code
- `task lint` - Run linting on all modules

### Protocol Buffers

- `task generate-proto` - Generate Go code from all proto files
- `task generate-proto-auth` - Generate auth service proto
- `task generate-proto-common` - Generate common proto messages
- `task clean-proto` - Clean generated proto files

### Building

- `task build` - Build all services
- `task build-service SERVICE=<name>` - Build specific service
- `task test` - Run tests on all modules

### Docker

- `task docker-build` - Build all Docker images
- `task docker-up` - Start all services with Docker Compose
- `task docker-down` - Stop all services

### Cleanup

- `task clean` - Clean all build artifacts and containers
- `task clean-build` - Clean build artifacts only
- `task clean-docker` - Stop containers and remove volumes

## Services

### Gateway Service (Port 8080)
HTTP API Gateway that routes requests to appropriate microservices.

### Auth Service (Port 9090)
gRPC authentication service providing user management and JWT tokens.

## Common Packages

### pkg/common
Shared utilities including:
- JWT authentication
- Structured logging
- HTTP middleware
- Redis client
- Database utilities
- Error handling
- Response formatting

### pkg/proto
Protocol buffer definitions for inter-service communication.

## Development Workflow

1. **Make changes to proto files** → `task generate-proto`
2. **Build and test** → `task build && task test`
3. **Format code** → `task format`
4. **Run locally** → `task docker-up`
5. **Clean up** → `task clean`
