-- AI Gateway 数据库表结构
create database ai_gateway charset utf8mb4;

use ai_gateway;

CREATE TABLE `abilities` (
     `group` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
     `model` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
     `channel_id` bigint NOT NULL,
     `enabled` tinyint(1) DEFAULT NULL,
     `priority` bigint DEFAULT '0',
     PRIMARY KEY (`group`,`model`,`channel_id`),
     KEY `idx_abilities_channel_id` (`channel_id`),
     KEY `idx_abilities_priority` (`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `channels` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `type` bigint DEFAULT '0',
    `key` text COLLATE utf8mb4_unicode_ci,
    `status` bigint DEFAULT '1',
    `name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `weight` bigint unsigned DEFAULT '0',
    `created_time` bigint DEFAULT NULL,
    `test_time` bigint DEFAULT NULL,
    `response_time` bigint DEFAULT NULL,
    `base_url` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '',
    `other` longtext COLLATE utf8mb4_unicode_ci,
    `balance` double DEFAULT NULL,
    `balance_updated_time` bigint DEFAULT NULL,
    `models` longtext COLLATE utf8mb4_unicode_ci,
    `group` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT 'default',
    `used_quota` bigint DEFAULT '0',
    `model_mapping` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT '',
    `priority` bigint DEFAULT '0',
    `config` longtext COLLATE utf8mb4_unicode_ci,
    `system_prompt` text COLLATE utf8mb4_unicode_ci,
    PRIMARY KEY (`id`),
    KEY `idx_channels_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `logs` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `user_id` bigint DEFAULT NULL,
    `created_at` bigint DEFAULT NULL,
    `type` bigint DEFAULT NULL,
    `content` longtext COLLATE utf8mb4_unicode_ci,
    `username` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '',
    `token_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '',
    `model_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '',
    `quota` bigint DEFAULT '0',
    `prompt_tokens` bigint DEFAULT '0',
    `completion_tokens` bigint DEFAULT '0',
    `channel_id` bigint DEFAULT NULL,
    `request_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '',
    `elapsed_time` bigint DEFAULT '0',
    `is_stream` tinyint(1) DEFAULT '0',
    `system_prompt_reset` tinyint(1) DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_logs_channel_id` (`channel_id`),
    KEY `idx_logs_user_id` (`user_id`),
    KEY `idx_created_at_type` (`created_at`,`type`),
    KEY `index_username_model_name` (`model_name`,`username`),
    KEY `idx_logs_token_name` (`token_name`),
    KEY `idx_logs_model_name` (`model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `options` (
    `key` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
    `value` longtext COLLATE utf8mb4_unicode_ci,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `redemptions` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `user_id` bigint DEFAULT NULL,
    `key` char(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `status` bigint DEFAULT '1',
    `name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `quota` bigint DEFAULT '100',
    `created_time` bigint DEFAULT NULL,
    `redeemed_time` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_redemptions_key` (`key`),
    KEY `idx_redemptions_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `tokens` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `user_id` bigint DEFAULT NULL,
    `key` char(48) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `status` bigint DEFAULT '1',
    `name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `created_time` bigint DEFAULT NULL,
    `accessed_time` bigint DEFAULT NULL,
    `expired_time` bigint DEFAULT '-1',
    `remain_quota` bigint DEFAULT '0',
    `unlimited_quota` tinyint(1) DEFAULT '0',
    `used_quota` bigint DEFAULT '0',
    `models` text COLLATE utf8mb4_unicode_ci,
    `subnet` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_tokens_key` (`key`),
    KEY `idx_tokens_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `users` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `password` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `display_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `role` bigint DEFAULT '1',
    `status` bigint DEFAULT '1',
    `email` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `github_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `wechat_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `lark_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `oidc_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `access_token` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `quota` bigint DEFAULT '0',
    `used_quota` bigint DEFAULT '0',
    `request_count` bigint DEFAULT '0',
    `group` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'default',
    `aff_code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `inviter_id` bigint DEFAULT NULL,
    `deleted_at` bigint DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_users_access_token` (`access_token`),
    UNIQUE KEY `idx_users_aff_code` (`aff_code`),
    UNIQUE KEY `uni_users_username` (`username`),
    KEY `idx_users_lark_id` (`lark_id`),
    KEY `idx_users_oidc_id` (`oidc_id`),
    KEY `idx_users_git_hub_id` (`github_id`),
    KEY `idx_users_inviter_id` (`inviter_id`),
    KEY `idx_users_username` (`username`),
    KEY `idx_users_display_name` (`display_name`),
    KEY `idx_users_email` (`email`),
    KEY `idx_users_we_chat_id` (`wechat_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
