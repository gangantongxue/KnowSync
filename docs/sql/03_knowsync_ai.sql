-- KnowSync AI 服务数据库 DDL
-- Database: knowsync_ai

CREATE DATABASE IF NOT EXISTS knowsync_ai
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE knowsync_ai;

-- AI 会话表
CREATE TABLE `chat_session` (
  `id`                 char(20)    NOT NULL COMMENT '会话 ID（xid）',
  `user_id`            varchar(20) NOT NULL COMMENT '用户 ID',
  `title`              varchar(255) NOT NULL DEFAULT '新对话' COMMENT '会话标题',
  `summary`            longtext    COMMENT '会话摘要（上下文压缩时生成）',
  `summary_updated_at` datetime    DEFAULT NULL COMMENT '摘要最后更新时间',
  `created_at`         datetime    NOT NULL COMMENT '创建时间',
  `updated_at`         datetime    NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  INDEX `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 会话表';

-- AI 消息表
CREATE TABLE `chat_message` (
  `id`         char(20)    NOT NULL COMMENT '消息 ID（xid）',
  `session_id` char(20)    NOT NULL COMMENT '所属会话 ID',
  `role`       varchar(16) NOT NULL COMMENT '角色：user / assistant',
  `content`    longtext    NOT NULL COMMENT '消息内容',
  `thinking`   longtext    COMMENT 'AI 思考过程（仅 assistant 角色）',
  `created_at` datetime    NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_session_id` (`session_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 消息表';
