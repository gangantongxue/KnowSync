# KnowSync 项目规则

## 日志规范

项目通过 `slog.SetDefault` 设置全局默认日志记录器。

- **所有业务代码**直接使用 `slog.Info` / `slog.Error` / `slog.Warn` / `slog.Debug` 等全局函数，不需要通过注入的日志对象调用
- 依赖注入中保留的 `Logger` 字段仅用于基础设施集成（如 GORM、Redis 等第三方库的日志适配器），业务层不应调用注入的日志对象
