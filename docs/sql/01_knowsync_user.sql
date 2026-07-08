-- KnowSync 用户服务数据库 DDL
-- Database: knowsync_user

CREATE DATABASE IF NOT EXISTS knowsync_user
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE knowsync_user;

-- 用户表
CREATE TABLE `user` (
  `id`         varchar(20)  NOT NULL COMMENT '用户 ID（数字自增转字符串）',
  `name`       varchar(64)  NOT NULL COMMENT '用户名',
  `email`      varchar(254) NOT NULL COMMENT '邮箱地址（登录凭证）',
  `password`   varchar(255) NOT NULL COMMENT 'bcrypt 加密密码',
  `avatar`     varchar(255) DEFAULT NULL COMMENT '头像 URL',
  `created_at` datetime     NOT NULL COMMENT '创建时间',
  `updated_at` datetime     NOT NULL COMMENT '更新时间',
  `deleted_at` datetime     DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_user_email` (`email`),
  INDEX `idx_user_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户会话表
CREATE TABLE `user_session` (
  `id`                  char(20)     NOT NULL COMMENT '会话 ID（xid）',
  `user_id`             varchar(20)  NOT NULL COMMENT '用户 ID',
  `refresh_token_hash`  varchar(255) NOT NULL COMMENT 'Refresh Token 哈希值',
  `client_ip`           varchar(64)  NOT NULL COMMENT '登录客户端 IP',
  `expire_at`           datetime     NOT NULL COMMENT '会话过期时间',
  `login_at`            datetime     NOT NULL COMMENT '登录时间',
  `logout_at`           datetime     DEFAULT NULL COMMENT '登出时间',
  `created_at`          datetime     NOT NULL COMMENT '创建时间',
  `updated_at`          datetime     NOT NULL COMMENT '更新时间',
  `deleted_at`          datetime     DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `user_id_refresh_token_hash_index` (`user_id`, `refresh_token_hash`),
  INDEX `idx_user_session_user_id` (`user_id`),
  INDEX `idx_user_session_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';
