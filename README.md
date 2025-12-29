# Scaffolding-Code-Generation

基于 Go 语言的 DDD 项目脚手架生成器。

## 功能特性

- 🚀 交互式命令行界面
- 📦 生成完整的 DDD 项目骨架
- 🔧 Go Workspace + BOM 依赖管理
- 🐳 Docker + PostgreSQL + Redis 配置
- ✨ 开箱即用的示例代码

## 安装

```bash
# 从源码构建
go build -o archi-gen ./cmd/archi-gen

# 或者直接运行
go run ./cmd/archi-gen init
```

## 使用方法

```bash
# 初始化新项目
archi-gen init
```

### 交互式流程

```
$ archi-gen init

🚀 欢迎使用 Archi-Gen 项目脚手架!

? 请输入项目名称: my-project
? 请选择开发语言: Go
? 请输入 Go 模块路径: github.com/yourname/my-project
? 是否使用 Redis? Yes

📋 项目配置:
   项目名称: my-project
   模块路径: github.com/yourname/my-project
   开发语言: go
   数据库:   PostgreSQL
   缓存:     Redis (是)
   部署方式: Docker

✨ 正在生成项目骨架...
   ✔ 创建项目目录
   ✔ 生成 go.work
   ✔ 生成 .gitignore
   ✔ 生成 Makefile
   ✔ 生成 BOM 模块
   ✔ 生成 share 模块
   ✔ 生成 user/domain 模块
   ✔ 生成 user/infrastructure 模块
   ✔ 生成 user 聚合模块
   ✔ 生成 api/user-api 模块
   ✔ 生成 api 聚合模块
   ✔ 生成 cmd/api 入口
   ✔ 生成 Dockerfile
   ✔ 生成 docker-compose.yml
   ✔ 生成 README.md

🎉 项目骨架生成成功!

📦 项目路径: ./my-project

🚀 快速开始:
   cd my-project
   go work sync
   docker-compose up -d postgres redis
   go run ./cmd/api/main.go
```

## 生成的项目结构

```
my-project/
├── go.work                   # Go 工作区配置
├── bom/                      # BOM 依赖管理模块
│   ├── go.mod
│   └── bom.go
├── share/                    # 公共组件模块
│   ├── go.mod
│   ├── errors/               # 错误定义
│   ├── utils/                # 工具函数
│   ├── types/                # 通用类型
│   └── middleware/           # 中间件
├── user/                     # 用户聚合模块
│   ├── go.mod
│   ├── domain/               # 领域层
│   │   ├── go.mod
│   │   ├── entity/           # 领域实体
│   │   ├── repository/       # 仓储接口
│   │   ├── service/          # 领域服务
│   │   ├── valueobject/      # 值对象
│   │   └── event/            # 领域事件
│   └── infrastructure/       # 基础设施层
│       ├── go.mod
│       ├── entity/           # 数据库实体 (PO)
│       ├── converter/        # 转换器
│       └── repository/       # 仓储实现
├── api/                      # API 聚合模块
│   ├── go.mod
│   └── user-api/
│       ├── go.mod
│       ├── dto/              # 数据传输对象
│       ├── service/          # 应用服务
│       └── http/             # HTTP 处理器
├── cmd/
│   └── api/                  # 主程序入口
│       ├── go.mod
│       └── main.go
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## 技术栈

- **语言**: Go 1.24+
- **CLI 框架**: [Cobra](https://github.com/spf13/cobra)
- **交互式提示**: [Survey](https://github.com/AlecAivazis/survey)

### 生成的项目技术栈

- **HTTP 框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **缓存**: Redis (可选)
- **容器化**: Docker

## 开发

```bash
# 安装依赖
go mod tidy

# 构建
go build -o bin/archi-gen ./cmd/archi-gen

# 运行测试
go test ./...
```

## License

MIT