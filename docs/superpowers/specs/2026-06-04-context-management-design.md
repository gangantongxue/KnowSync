# LLM 对话上下文管理设计

## 背景

当前 `buildMessages()` 将对话的完整消息历史全部发送给 LLM，不做任何截断或压缩。随着对话增长，总 token 数可能超过模型上下文窗口（DeepSeek Chat 128K），导致 API 静默截断或报错。

## 目标

为 knowsync 的 AI 知识库问答助手添加智能上下文管理机制，类似 opencode 的 `compaction` 策略：

1. 自动检测上下文是否即将溢出
2. 通过裁剪旧工具输出释放 token
3. 必要时将早期对话摘要压缩为结构化摘要，保留关键信息
4. 全自动运行，无需用户干预

## 架构

### 新增组件

`ai-server/internal/service/compactor/` 包，包含三个模块：

| 文件 | 职责 |
|------|------|
| `compactor.go` | 编排器：溢出检测 → 裁剪 → 压缩 |
| `summarizer.go` | 将早期消息发送给 LLM 生成结构化摘要 |
| `tokenizer.go` | 基于字符的 token 快速估算 |

### 修改组件

| 文件 | 变更 |
|------|------|
| `ai-server/internal/service/chat.go` | `Chat()` 中集成 Compactor |
| `ai-server/internal/repository/session.go` | 新增 summary 字段读写 |
| `ai-server/internal/repository/repository.go` | AutoMigrate 新字段 |

### 数据模型变更

`ChatSession` 表新增字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `summary` | `longtext` | 结构化摘要内容（Markdown 格式），可为空 |
| `summary_updated_at` | `datetime` | 最近一次摘要更新时间，用于并发控制 |

## 核心流程

```
Chat(ctx, req)
  → getOrCreateSession()
  → CreateMessage(userMsg)
  → msgs = buildMessages(sessionID)
  → msgs = Compactor.CompactIfNeeded(ctx, msgs, session)
  → LLM.Stream(msgs)
  → CreateMessage(assistantMsg)
  → Compactor.RecordUsage(session, usage)   // 校准 token 统计
```

## 溢出检测

### Token 估算

使用字符/token 经验比例（4 字符 ≈ 1 token）：

```go
const avgCharPerToken = 4

func EstimateTokens(text string) int {
    return int(math.Ceil(float64(len(text)) / avgCharPerToken))
}
```

针对 LLM 消息格式，每条消息额外 +10 token 的角色/格式开销。

### 阈值

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `ModelContextLimit` | 128_000 | DeepSeek Chat 上下文窗口 |
| `ReservedTokens` | 20_000 | 保留给模型输出的 buffer |
| `OverflowRatio` | 0.8 | 达到上下文 80% 时触发预压缩 |

触发条件：`totalTokens >= (ModelContextLimit - ReservedTokens) × OverflowRatio`

### 精确校准

每次 LLM 调用后，从 API 响应中解析 `usage.input_tokens` / `output_tokens`，记录到 `ChatSession`，后续使用实际值替代估算值判断。

## 裁剪 (Prune)

在触发压缩时优先执行，因为成本最低。

1. 从最早的消息开始向后遍历
2. 跳过最近 `PruneProtectedTurns`（默认 2）轮中的工具输出
3. 对于更早的工具输出，若内容长度 > `PruneToolOutputMinChars`（默认 2000），替换为截断标记：`[工具输出已裁剪，完整内容请参考数据库记录]`
4. 若裁剪释放的 token > `PruneTargetRelease`（默认 20K），跳过摘要压缩

## 摘要压缩 (Compact)

### 保留内容

- 最近 `CompactionKeepRecentTurns`（默认 3）轮保持完整
- 若已有旧摘要，保留它
- 其余早期消息送入摘要生成

### 摘要生成

调用 LLM（与主模型相同），使用结构化提示词生成摘要：

```
请将以下对话历史压缩为结构化摘要，保留关键信息：

## 用户核心意图
[用户最初想解决什么问题]

## 已讨论的关键点
- [重要的事实、决策、确认]

## 已获取的知识库信息
- [从知识库检索到的关键内容]

## 待办/未完成
- [用户提到但尚未完成的事项]
```

**增量更新**：如果已有旧摘要，将旧摘要 + 后续未摘要的消息一起传给 LLM，要求更新而不是重写。

### 摘要存储

生成的摘要写入 `ChatSession.summary`。
使用 `summary_updated_at` 做乐观锁，避免并发覆盖。

### 最终消息组装

```
compacted = []
if session.Summary != "":
    compacted += [SystemMessage(session.Summary)]  // 作为 system 消息
compacted += recentMessages                        // 最近 N 轮完整消息
```

## Token 用量记录

每次 LLM 调用完成后，解析流式响应中的 usage 元数据：

```go
type Usage struct {
    InputTokens  int
    OutputTokens int
}
```

将累积 token 数暂存在会话内存对象中（不持久化到 DB，避免频繁写入），仅用于后续溢出检测的准确判断。若 API 未返回 usage 信息，继续使用字符估算值。

## 边界情况

| 场景 | 处理 |
|------|------|
| 首次对话 | 跳过压缩 |
| 对话 < 3 轮 | 跳过压缩 |
| 压缩后仍超限 | 硬截断：从最早的非摘要消息开始逐条丢弃，每次丢弃后重新估算，直到低于阈值；仍不行则返回 ContextOverflowError |
| LLM 摘要调用失败 | 回退到仅裁剪模式 |
| API 无 usage 返回 | 继续使用字符估算 |
| 并发请求同一会话 | `summary_updated_at` 乐观锁 |

## 默认配置（硬编码）

```go
const (
    OverflowRatio            = 0.8
    ReservedTokens           = 20_000
    PruneToolOutputMinChars  = 2_000
    PruneTargetRelease       = 20_000
    PruneProtectedTurns      = 2
    CompactionKeepRecentTurns = 3
    CompactionMaxSummaryLen  = 2_000
)
```

## 未纳入范围

- 不支持用户自定义配置（全自动）
- 不支持切换摘要模型（复用主模型）
- 不涉及前端改动
