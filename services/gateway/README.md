# Envoy-Only API Gateway

This directory contains the Envoy proxy configuration that serves as the API gateway for the microservices architecture.

## Architecture

```
Client → Envoy (port 8080) → Auth Service (HTTP 8081 / gRPC 9090)
```

## Why Envoy-Only?

- **Industry Standard**: Used by Google, Netflix, Uber, and other major companies
- **High Performance**: C++ based, extremely fast and efficient
- **Rich Features**: Load balancing, health checks, circuit breaking, retries, observability
- **Service Mesh Ready**: Easy to extend to Istio or Consul Connect later
- **Less Code to Maintain**: No custom gateway code needed

## Configuration

The main configuration is in [`config/envoy.yaml`](config/envoy.yaml):

### Key Features
- **Direct Routing**: Routes API calls directly to auth service
- **Health Checks**: Automatic monitoring of backend services  
- **CORS Support**: Configurable CORS policies
- **gRPC-Web**: Support for browser gRPC clients
- **Admin Interface**: Monitoring and statistics on port 9901
- **Retry Logic**: Automatic retries for failed requests
- **Load Balancing**: Round-robin between service instances

### Routes
- `POST /api/v1/auth/*` → Auth Service HTTP (port 8081)
- `GET /health` → Auth Service health check
- `/auth.v1.AuthService/*` → Auth Service gRPC (port 9090) for gRPC-Web clients

## Admin Interface

Access Envoy's admin interface at `http://localhost:9901`:

- `/stats` - Real-time statistics
- `/clusters` - Backend service health
- `/config_dump` - Current configuration
- `/ready` - Readiness check

## Files

- `Dockerfile.envoy` - Container build configuration
- `config/envoy.yaml` - Main Envoy configuration
- ~~`Dockerfile`~~ - *Removed* (was for Go gateway)
- ~~`cmd/`~~ - *Removed* (was for Go gateway code)
- ~~`internal/`~~ - *Removed* (was for Go gateway code)

## Benefits of This Approach

1. **Simplified Architecture**: One less service to manage
2. **Better Performance**: Direct routing without intermediate proxy
3. **Industry Best Practices**: Using proven enterprise-grade proxy
4. **Observability**: Built-in metrics and monitoring
5. **Scalability**: Easy to add more backend services

## Testing

```powershell
# Test authentication through Envoy
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -ContentType "application/json" -Body '{"email":"test@example.com","password":"password123"}'

# Check Envoy stats
Invoke-RestMethod -Uri "http://localhost:9901/stats" -Method Get
```

## Adding New Services

To add a new microservice to the gateway:

1. **Add cluster configuration** in `config/envoy.yaml`:
```yaml
clusters:
- name: product_service_http
  connect_timeout: 0.25s
  type: LOGICAL_DNS
  lb_policy: ROUND_ROBIN
  load_assignment:
    cluster_name: product_service_http
    endpoints:
    - lb_endpoints:
      - endpoint:
          address:
            socket_address:
              address: product  # Service name in docker-compose
              port_value: 8082
  health_checks:
  - timeout: 1s
    interval: 10s
    unhealthy_threshold: 3
    healthy_threshold: 3
    http_health_check:
      path: "/health"
```

2. **Add route configuration**:
```yaml
virtual_hosts:
- name: local_service
  domains: ["*"]
  routes:
  - match:
      prefix: "/api/v1/products"
    route:
      cluster: product_service_http
```

3. **Add service to docker-compose.yml**:
```yaml
product:
  build: ./services/product
  ports:
    - "8082:8082"
  depends_on:
    - postgres
    - redis
```

This configuration provides a production-ready API gateway using industry-standard technology.

## Production Considerations

### Security
- Configure proper CORS origins (replace `*` with actual domains)
- Use HTTPS/TLS in production
- Set up rate limiting and DDoS protection
- Enable request/response size limits

### Performance
- Tune connection pool settings
- Configure appropriate timeouts
- Monitor memory and CPU usage
- Enable HTTP/2 and gRPC-Web compression

### Monitoring
- Set up Prometheus metrics collection
- Configure custom dashboards in Grafana
- Enable distributed tracing with Jaeger
- Set up alerting for service health

### High Availability
- Deploy multiple Envoy instances behind a load balancer
- Configure proper health checks
- Set up circuit breakers for failing services
- Enable automatic retries and failover
