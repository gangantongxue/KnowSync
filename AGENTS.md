# Knowsync

Go 微服务项目，基于 gRPC 通信（user-server/repo-server/ai-server），HTTP 网关层（gateway）使用 Hertz 框架。

## 架构规范

- **严格遵守 3 层架构**：handler（gRPC/HTTP 处理）→ service（业务逻辑+权限校验）→ repository（纯数据访问）。禁止跨层调用、禁止业务逻辑侵入 handler 或 repository 层
- **禁止 JOIN 查询**：多表关联时执行多次独立查询，在 service 层组合数据
- **禁止数据库外键**：字段保留逻辑引用关系（如 `repo_id`），但不设置 `FOREIGN KEY` 约束
- **gateway 无 service 层**：gateway handler 直接通过 gRPC client 调用后端服务，不封装业务逻辑

## 代码风格一致性

新增代码须与已有代码风格保持一致，包括但不限于：

- **导入分组**：标准库 → 第三方包（含 `ks-proto`）→ 项目内部包，每组间空行分隔
- **错误处理**：handler 层返回 `{Success: false, Msg: err.Error()}, nil`（Go error 不向上透传）；service 层用 `fmt.Errorf("中文描述: %w", err)` 包裹并 `slog.Error` 记录；repository 层透传原始 error
- **结构体标签**：GORM model 同时包含 `gorm` 和 `json` 标签；配置结构体同时包含 `yaml` 和 `mapstructure` 标签
- **注释**：所有导出结构体、函数/方法均需中文注释，使用 `// ` 单行格式
- **构造器**：使用 `NewXxx(...) (*Xxx, error)` 模式，依赖通过参数注入，不使用 `samber/do` 等 DI 容器

## 日志规范

项目通过 `slog.SetDefault` 设置全局默认日志记录器。

- **所有业务代码**直接使用 `slog.Info` / `slog.Error` / `slog.Warn` / `slog.Debug` 等全局函数，不需要通过注入的日志对象调用
- 依赖注入中保留的 `Logger` 字段仅用于基础设施集成（如 GORM、Redis 等第三方库的日志适配器），业务层不应调用注入的日志对象

## 配置管理规范

- **使用 viper 自动解析** — 禁止逐字段手动绑定，统一使用 `v.AutomaticEnv()` + `v.Unmarshal(&cfg)` 自动解析配置
- **所有配置结构体字段**须同时包含 `yaml` 和 `mapstructure` 标签，确保 YAML 反序列化与环境变量绑定一致
- **环境变量优先级高于配置文件** — 始终在 `v.ReadInConfig()` 之前调用 `v.AutomaticEnv()`，利用 viper 的 env > config 优先级顺序
- **环境变量前缀**遵循 `KNOWSYNC_{SERVER_NAME}` 格式（如 `KNOWSYNC_USER_SERVER`、`KNOWSYNC_GATEWAY`），键路径中的 `.` 替换为 `_`

## 执行命令规范

- 有可用的 `task` 命令，优先使用 `task` 命令执行任务代替直接执行 go 或其他命令
