-- KnowSync 聊天服务数据库 DDL
-- Database: knowsync_chat

CREATE DATABASE IF NOT EXISTS knowsync_chat
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE knowsync_chat;

-- 聊天用户缓存表（从 user-server 同步）
CREATE TABLE `users` (
  `id`       varchar(20)  NOT NULL COMMENT '用户 ID',
  `name`     varchar(50)  NOT NULL COMMENT '用户名',
  `email`    varchar(100) NOT NULL COMMENT '邮箱',
  `avatar`   varchar(255) DEFAULT '' COMMENT '头像 URL',
  `password` varchar(255) NOT NULL COMMENT '密码',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天用户缓存表';

-- 好友关系表
CREATE TABLE `friends` (
  `id`               char(20)        NOT NULL COMMENT '记录 ID（xid）',
  `user_id`          varchar(20)     NOT NULL COMMENT '用户 ID',
  `friend_id`        varchar(20)     NOT NULL COMMENT '好友用户 ID',
  `remark`           varchar(100)    DEFAULT '' COMMENT '好友备注',
  `last_message_at`  bigint          NOT NULL DEFAULT 0 COMMENT '最后消息时间戳',
  `last_read_seq_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '已读的最大消息 SeqID',
  `pinned`           tinyint(1)      NOT NULL DEFAULT 0 COMMENT '是否置顶：0 否 / 1 是',
  `created_at`       bigint          NOT NULL COMMENT '成为好友的时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uk_user_friend` (`user_id`, `friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

-- 好友申请表
CREATE TABLE `friend_requests` (
  `id`          char(20)    NOT NULL COMMENT '记录 ID（xid）',
  `sender_id`   varchar(20) NOT NULL COMMENT '发送方用户 ID',
  `receiver_id` varchar(20) NOT NULL COMMENT '接收方用户 ID',
  `status`      enum('pending','accepted','rejected') NOT NULL DEFAULT 'pending' COMMENT '申请状态',
  `remark`      varchar(100) DEFAULT '' COMMENT '申请附言',
  `created_at`  bigint      NOT NULL COMMENT '发送时间',
  `updated_at`  bigint      NOT NULL COMMENT '处理时间',
  PRIMARY KEY (`id`),
  INDEX `idx_sender` (`sender_id`),
  INDEX `idx_receiver_status` (`receiver_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友申请表';

-- 消息表
CREATE TABLE `messages` (
  `id`                char(20)    NOT NULL COMMENT '消息 ID（xid）',
  `conversation_type` enum('private','group') NOT NULL COMMENT '会话类型',
  `conversation_id`   varchar(64) NOT NULL COMMENT '会话 ID',
  `seq_id`            bigint unsigned NOT NULL COMMENT '会话内递增序列号',
  `sender_id`         varchar(20) NOT NULL COMMENT '发送方用户 ID',
  `content_type`      enum('text','image','file','system_invitation') NOT NULL DEFAULT 'text' COMMENT '消息内容类型',
  `content`           text        NOT NULL COMMENT '消息内容',
  `extra`             json        DEFAULT NULL COMMENT '扩展数据',
  `reply_to_id`       char(20)    DEFAULT NULL COMMENT '回复的消息 ID',
  `status`            enum('normal','recalled') NOT NULL DEFAULT 'normal' COMMENT '消息状态',
  `created_at`        bigint      NOT NULL COMMENT '发送时间',
  PRIMARY KEY (`id`),
  INDEX `idx_conversation` (`conversation_id`, `seq_id`),
  INDEX `idx_sender` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

-- 群组表
CREATE TABLE `groups` (
  `id`         char(20)     NOT NULL COMMENT '群组 ID（xid）',
  `name`       varchar(100) NOT NULL COMMENT '群组名称',
  `avatar`     varchar(500) DEFAULT '' COMMENT '群头像 URL',
  `owner_id`   varchar(20)  NOT NULL COMMENT '群主用户 ID',
  `created_at` bigint       NOT NULL COMMENT '创建时间',
  `updated_at` bigint       NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组表';

-- 群组成员表
CREATE TABLE `group_members` (
  `id`               char(20)        NOT NULL COMMENT '记录 ID（xid）',
  `group_id`         char(20)        NOT NULL COMMENT '群组 ID',
  `user_id`          varchar(20)     NOT NULL COMMENT '成员用户 ID',
  `role`             enum('owner','admin','member') NOT NULL DEFAULT 'member' COMMENT '群内角色',
  `last_read_seq_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '已读的最大消息 SeqID',
  `pinned`           tinyint(1)      NOT NULL DEFAULT 0 COMMENT '是否置顶',
  `joined_at`        bigint          NOT NULL COMMENT '加入时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uk_group_user` (`group_id`, `user_id`),
  INDEX `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组成员表';
