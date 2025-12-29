package generator

// generateWorkspace 生成 go.work 文件
func (g *GoGenerator) generateWorkspace() error {
	tmpl := `go 1.24.11

use (
	./bom
	./share
	./user
	./user/domain
	./user/infrastructure
	./api
	./api/user-api
	./cmd/api
)
`
	return g.writeFile("go.work", tmpl)
}

// generateGitignore 生成 .gitignore 文件
func (g *GoGenerator) generateGitignore() error {
	tmpl := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary
*.test

# Output of the go coverage tool
*.out
coverage.html
coverage.txt

# Dependency directories
vendor/

# Go workspace lock file
go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
*.local

# Logs
*.log
logs/

# Temp files
tmp/
temp/
`
	return g.writeFile(".gitignore", tmpl)
}

// generateMakefile 生成 Makefile
func (g *GoGenerator) generateMakefile() error {
	tmpl := `.PHONY: build run test clean tidy docker-up docker-down

# 构建
build:
	go build -o bin/api ./cmd/api

# 运行
run:
	go run ./cmd/api/main.go

# 测试
test:
	go test -v ./...

# 测试覆盖率
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# 同步依赖
tidy:
	cd bom && go mod tidy
	cd share && go mod tidy
	cd user/domain && go mod tidy
	cd user/infrastructure && go mod tidy
	cd user && go mod tidy
	cd api/user-api && go mod tidy
	cd api && go mod tidy
	cd cmd/api && go mod tidy
	go work sync

# 启动 Docker 服务
docker-up:
	docker-compose up -d

# 停止 Docker 服务
docker-down:
	docker-compose down

# 查看 Docker 日志
docker-logs:
	docker-compose logs -f

# 重新构建并启动
docker-rebuild:
	docker-compose up -d --build
`
	return g.writeFile("Makefile", tmpl)
}

// generateDockerfile 生成 Dockerfile
func (g *GoGenerator) generateDockerfile() error {
	tmpl := `# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.work and all module files
COPY go.work ./
COPY bom/go.mod ./bom/
COPY share/go.mod ./share/
COPY user/go.mod ./user/
COPY user/domain/go.mod ./user/domain/
COPY user/infrastructure/go.mod ./user/infrastructure/
COPY api/go.mod ./api/
COPY api/user-api/go.mod ./api/user-api/
COPY cmd/api/go.mod ./cmd/api/

# Download dependencies
RUN go work sync

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
`
	return g.writeFile("Dockerfile", tmpl)
}

// generateDockerCompose 生成 docker-compose.yml
func (g *GoGenerator) generateDockerCompose() error {
	tmpl := `version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: {{.ProjectName}}-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: {{.ProjectName}}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - {{.ProjectName}}-network
{{if .UseRedis}}
  redis:
    image: redis:7-alpine
    container_name: {{.ProjectName}}-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - {{.ProjectName}}-network
{{end}}
  app:
    build: .
    container_name: {{.ProjectName}}-app
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: {{.ProjectName}}
{{if .UseRedis}}      REDIS_HOST: redis
      REDIS_PORT: 6379
{{end}}    depends_on:
      postgres:
        condition: service_healthy
{{if .UseRedis}}      redis:
        condition: service_healthy
{{end}}    networks:
      - {{.ProjectName}}-network

volumes:
  postgres_data:
{{if .UseRedis}}  redis_data:
{{end}}
networks:
  {{.ProjectName}}-network:
    driver: bridge
`
	return g.renderAndWrite(tmpl, "docker-compose.yml")
}

// generateReadme 生成 README.md
func (g *GoGenerator) generateReadme() error {
	tmpl := `# {{.ProjectName}}

基于 Go 语言的领域驱动设计（DDD）项目，采用多模块工作区（Go Workspace）+ BOM 依赖管理。

## 技术栈

- **语言**: Go 1.24.11
- **HTTP 框架**: Hertz (CloudWeGo)
- **RPC 框架**: Kitex (CloudWeGo)
- **ORM**: GORM
- **数据库**: PostgreSQL 16
{{if .UseRedis}}- **缓存**: Redis 7{{end}}
- **依赖管理**: BOM (Bill of Materials)
- **容器化**: Docker + Docker Compose

## 快速开始

### 1. 同步依赖

` + "```bash" + `
go work sync
` + "```" + `

### 2. 启动数据库服务

` + "```bash" + `
docker-compose up -d postgres{{if .UseRedis}} redis{{end}}
` + "```" + `

### 3. 运行应用

` + "```bash" + `
go run ./cmd/api/main.go
` + "```" + `

访问 http://localhost:8080/health 检查服务状态。

## 项目结构

` + "```" + `
{{.ProjectName}}/
├── go.work                   # Go 工作区配置
├── bom/                      # BOM 依赖管理模块
├── share/                    # 公共组件模块
│   ├── errors/               # 错误定义
│   ├── utils/                # 工具函数
│   ├── types/                # 通用类型
│   └── middleware/           # 中间件
├── user/                     # 用户聚合模块
│   ├── domain/               # 领域层
│   │   ├── entity/           # 领域实体
│   │   ├── repository/       # 仓储接口
│   │   ├── service/          # 领域服务
│   │   ├── valueobject/      # 值对象
│   │   └── event/            # 领域事件
│   └── infrastructure/       # 基础设施层
│       ├── entity/           # 数据库实体 (PO)
│       ├── converter/        # 转换器
│       └── repository/       # 仓储实现
├── api/                      # API 聚合模块
│   └── user-api/             # 用户 API
│       ├── dto/              # 数据传输对象
│       ├── service/          # 应用服务
│       └── http/             # HTTP 处理器
└── cmd/
    └── api/                  # 主程序入口
` + "```" + `

## 环境变量

- ` + "`DB_HOST`" + `: PostgreSQL 主机（默认：localhost）
- ` + "`DB_PORT`" + `: PostgreSQL 端口（默认：5432）
- ` + "`DB_USER`" + `: 数据库用户（默认：postgres）
- ` + "`DB_PASSWORD`" + `: 数据库密码（默认：postgres）
- ` + "`DB_NAME`" + `: 数据库名称（默认：{{.ProjectName}}）
{{if .UseRedis}}- ` + "`REDIS_HOST`" + `: Redis 主机（默认：localhost）
- ` + "`REDIS_PORT`" + `: Redis 端口（默认：6379）{{end}}

## 常用命令

` + "```bash" + `
# 构建
make build

# 运行
make run

# 测试
make test

# 同步依赖
make tidy

# 启动 Docker 服务
make docker-up

# 停止 Docker 服务
make docker-down
` + "```" + `

## License

MIT
`
	return g.renderAndWrite(tmpl, "README.md")
}

// generateDockerignore 生成 .dockerignore 文件
func (g *GoGenerator) generateDockerignore() error {
	tmpl := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary
*.test

# IDE
.idea/
.vscode/
*.swp
*.swo

# Git
.git/
.gitignore

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
*.local

# Logs
*.log
logs/

# Docker
Dockerfile
docker-compose*.yml
.dockerignore

# Temp
tmp/
temp/

# Documentation
*.md
!README.md

# Tests
*_test.go
coverage.*
`
	return g.writeFile(".dockerignore", tmpl)
}
