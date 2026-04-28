# PostgreSQL 扩展镜像

本镜像基于 `pgvector/pgvector:pg18-trixie` 构建，并内置了多个常用的扩展插件。

## 包含的插件

- **pgvector**: 原生自带，提供向量相似度搜索能力。
- **pg_stat_statements**: 用于追踪 SQL 语句执行统计信息的扩展（已在配置中自动加入 `shared_preload_libraries`）。
- **zhparser**: 基于 SCWS (Simple Chinese Word Segmentation) 的 PostgreSQL 中文分词扩展。
- **pgmq**: 基于 PostgreSQL 的轻量级消息队列。

## 环境变量

除了原版 PostgreSQL 的所有环境变量（如 `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`）外，没有任何特殊定制。

## 镜像地址

`ghcr.io/yorkane/postgres:latest`

## 使用方式

可以直接在 `docker-compose.yml` 中引用此镜像：

```yaml
services:
  postgres:
    image: ghcr.io/yorkane/postgres:latest
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: mysecretpassword
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
    volumes:
      - pg-data:/var/lib/postgresql/data

volumes:
  pg-data:
```

在容器首次启动初始化数据库时，以下扩展将会被自动启用：

- `pg_stat_statements`
- `vector`
- `pg_trgm`
- `zhparser`
- `pgmq`

无需手动进入数据库执行 `CREATE EXTENSION` 命令。