# Docker Guide

This guide explains Docker concepts and how our microservices are containerized.

## What is Docker?

Docker is a platform for developing, shipping, and running applications using containers. Containers package an application with all its dependencies, ensuring it runs consistently across different environments.

### Key Concepts

- **Image**: A read-only template for creating containers
- **Container**: A running instance of an image
- **Dockerfile**: Instructions for building an image
- **Registry**: A repository for storing images (e.g., Docker Hub, ECR)

## Multi-Stage Builds

Our Dockerfiles use **multi-stage builds** to create smaller, more secure images.

### Why Multi-Stage Builds?

1. **Smaller Images**: Only include runtime dependencies, not build tools
2. **Security**: Fewer components mean smaller attack surface
3. **Performance**: Faster image pulls and container starts

### Build Stages

#### Stage 1: Builder
```dockerfile
FROM golang:1.23-alpine AS builder
```
- Uses full Go toolchain
- Compiles the application
- Creates statically linked binary

#### Stage 2: Runtime
```dockerfile
FROM alpine:latest
```
- Minimal base image (~5MB)
- Only includes the compiled binary
- Runtime dependencies (CA certificates, timezone data)

## Dockerfile Structure

### Example: Auth Service Dockerfile

```dockerfile
# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/auth-service .
USER 1000:1000
EXPOSE 8080
CMD ["./auth-service"]
```

### Key Instructions

- **FROM**: Base image for the stage
- **WORKDIR**: Set working directory
- **COPY**: Copy files into image
- **RUN**: Execute commands during build
- **USER**: Switch to non-root user (security)
- **EXPOSE**: Document which port the app uses
- **CMD**: Default command to run

## Security Best Practices

### 1. Non-Root User

Always run containers as non-root:

```dockerfile
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
USER appuser
```

### 2. Minimal Base Images

Use Alpine Linux instead of full Linux distributions:

```dockerfile
FROM alpine:latest  # ~5MB vs ~100MB+
```

### 3. Layer Caching

Order Dockerfile instructions to maximize cache hits:

```dockerfile
# Copy dependency files first (changes less frequently)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code last (changes frequently)
COPY . .
```

### 4. Health Checks

Add health checks for container orchestration:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --spider http://localhost:8080/health || exit 1
```

## Building Images

### Build Command

```bash
docker build -f docker/Dockerfile.auth-service -t auth-service:latest .
```

### Build Arguments

```bash
docker build \
  --build-arg VERSION=1.0.0 \
  -t auth-service:1.0.0 \
  -f docker/Dockerfile.auth-service .
```

## Image Optimization

### 1. Use .dockerignore

Create `.dockerignore` to exclude unnecessary files:

```
.git
.vscode
*.md
test/
.env
```

### 2. Minimize Layers

Combine RUN commands:

```dockerfile
# Bad
RUN apk update
RUN apk add git
RUN apk add curl

# Good
RUN apk update && \
    apk add git curl && \
    rm -rf /var/cache/apk/*
```

### 3. Use Specific Tags

Avoid `latest` tag in production:

```dockerfile
FROM golang:1.23-alpine  # Specific version
```

## Pushing to ECR

### 1. Authenticate

```bash
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin <ECR_URL>
```

### 2. Tag Image

```bash
docker tag auth-service:latest <ECR_URL>/auth-service:latest
```

### 3. Push Image

```bash
docker push <ECR_URL>/auth-service:latest
```

## Image Sizes

Our optimized images:

- **auth-service**: ~15MB
- **expense-service**: ~15MB
- **receipt-service**: ~15MB
- **notification-service**: ~12MB

Compare to non-optimized: ~500MB+ (with full Go toolchain)

## Troubleshooting

### Build Fails

1. Check Dockerfile syntax
2. Verify base image exists
3. Check network connectivity (for `go mod download`)

### Image Too Large

1. Use multi-stage builds
2. Remove unnecessary files
3. Use `.dockerignore`
4. Use Alpine base images

### Container Won't Start

1. Check CMD/ENTRYPOINT
2. Verify exposed ports
3. Check file permissions
4. Review container logs: `docker logs <container-id>`

## Best Practices Summary

✅ Use multi-stage builds  
✅ Run as non-root user  
✅ Use minimal base images  
✅ Leverage layer caching  
✅ Add health checks  
✅ Use specific image tags  
✅ Scan images for vulnerabilities  
✅ Keep images up to date  

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Best Practices Guide](https://docs.docker.com/develop/dev-best-practices/)
- [Security Scanning](https://docs.docker.com/docker-hub/vulnerability-scanning/)
