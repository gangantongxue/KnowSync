## 日志规范

项目通过 `slog.SetDefault` 设置全局默认日志记录器。

- **所有业务代码**直接使用 `slog.Info` / `slog.Error` / `slog.Warn` / `slog.Debug` 等全局函数，不需要通过注入的日志对象调用
- 依赖注入中保留的 `Logger` 字段仅用于基础设施集成（如 GORM、Redis 等第三方库的日志适配器），业务层不应调用注入的日志对象

## 数据库设计规范

- **禁止使用数据库外键** — 字段保留逻辑引用关系（如 `repo_id`、`user_id`），但不在 MySQL 中设置 `FOREIGN KEY` 约束
- **禁止 JOIN 查询** — 涉及多表关联时，执行多次独立查询，在应用层组合数据

## 代码注释规范

- **所有 Go 文件**中的结构体、函数/方法定义、关键逻辑均需添加中文注释
- 注释风格参考项目现有代码（`user-server/` 中的注释模式），使用 `// ` 单行注释

## 执行命令规范

- 有可用的 `task` 命令，优先使用 `task` 命令执行任务

## 配置管理规范

- **使用 viper 自动解析** — 禁止逐字段手动绑定，统一使用 `v.AutomaticEnv()` + `v.Unmarshal(&cfg)` 自动解析配置
- **所有配置结构体字段**须同时包含 `yaml` 和 `mapstructure` 标签，确保 YAML 反序列化与环境变量绑定一致
- **环境变量优先级高于配置文件** — 始终在 `v.ReadInConfig()` 之前调用 `v.AutomaticEnv()`，利用 viper 的 env > config 优先级顺序
- **环境变量前缀**遵循 `KNOWSYNC_{SERVER_NAME}` 格式（如 `KNOWSYNC_USER_SERVER`、`KNOWSYNC_GATEWAY`），键路径中的 `.` 替换为 `_`
