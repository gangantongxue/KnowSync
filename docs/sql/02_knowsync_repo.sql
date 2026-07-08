-- KnowSync 知识库服务数据库 DDL
-- Database: knowsync_repo

CREATE DATABASE IF NOT EXISTS knowsync_repo
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE knowsync_repo;

-- 知识库表
CREATE TABLE `repo` (
  `id`            char(20)     NOT NULL COMMENT '知识库 ID（xid）',
  `owner_id`      varchar(20)  NOT NULL COMMENT '所有者用户 ID',
  `name`          varchar(128) NOT NULL COMMENT '知识库名称',
  `visibility`    varchar(20)  NOT NULL DEFAULT 'PRIVATE' COMMENT '可见性：PUBLIC / PRIVATE',
  `description`   varchar(500) DEFAULT NULL COMMENT '知识库描述',
  `article_count` bigint       NOT NULL DEFAULT 0 COMMENT '文章数量（冗余计数）',
  `created_at`    datetime     NOT NULL COMMENT '创建时间',
  `updated_at`    datetime     NOT NULL COMMENT '更新时间',
  `deleted_at`    datetime     DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  INDEX `idx_repo_owner_id` (`owner_id`),
  INDEX `idx_repo_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库表';

-- 协作者表
CREATE TABLE `collaborator` (
  `id`      char(20)    NOT NULL COMMENT '记录 ID（xid）',
  `repo_id` char(20)    NOT NULL COMMENT '知识库 ID',
  `user_id` varchar(20) NOT NULL COMMENT '协作者用户 ID',
  `role`    varchar(20) NOT NULL COMMENT '角色：ADMIN / DEVELOPER / VIEWER',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_repo_user` (`repo_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='协作者表';

-- 关注表
CREATE TABLE `follow` (
  `id`         char(20)    NOT NULL COMMENT '记录 ID（xid）',
  `user_id`    varchar(20) NOT NULL COMMENT '用户 ID',
  `repo_id`    char(20)    NOT NULL COMMENT '被关注的知识库 ID',
  `created_at` datetime    NOT NULL COMMENT '关注时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_user_repo` (`user_id`, `repo_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关注表';
