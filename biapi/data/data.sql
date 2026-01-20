CREATE TABLE `bi_template` (
    `template_id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板ID',
    `platform_id` INT UNSIGNED NOT NULL COMMENT '平台ID',
    `code` VARCHAR(64) NOT NULL COMMENT '模板业务编码',
    `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '模板名称',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '0=草稿 1=启用 2=禁用 3=废弃',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间，NULL=未删除',
    PRIMARY KEY (`template_id`) USING BTREE,
    -- 同一平台下，业务编码唯一（软删不冲突）
    UNIQUE KEY `uk_platform_code` (`platform_id`, `code`, `deleted_at`),
    -- 平台 + 状态查询优化
    KEY `idx_platform_status` (`platform_id`, `status`, `deleted_at`)
) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci
COMMENT='BI 模板定义表';




CREATE TABLE `bi_template_data` (
    `td_id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板数据ID',
    `platform_id` INT UNSIGNED NOT NULL COMMENT '平台ID',
    `template_id` BIGINT UNSIGNED NOT NULL COMMENT '模板ID',
    `company_id` BIGINT UNSIGNED NOT NULL COMMENT '公司ID（冗余，便于隔离）',
     env ENUM('test','gray','prod') NOT NULL DEFAULT 'test' COMMENT '环境：test,gray,prod',
    `op_type` SMALLINT UNSIGNED NOT NULL COMMENT '操作类型：201=add 202=update 203=del 401=list 402=detail 403=count',
    `content` MEDIUMTEXT NOT NULL COMMENT 'SQL / DSL 内容',
    `checksum` CHAR(32) NOT NULL DEFAULT '' COMMENT 'content 的 md5，用于缓存和变更检测',
    `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0=禁用 1=启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间，NULL=未删除',
    PRIMARY KEY (`td_id`) USING BTREE,
    -- 同一模板 + 公司 + 环境 + 操作类型 唯一
    UNIQUE KEY `uk_tpl_op` (`template_id`, `company_id`, `platform_id`, `env`, `op_type`, `deleted_at`),
    -- 模板 + 环境查询
    KEY `idx_tpl_env` (`template_id`, `env`, `deleted_at`),
    -- 公司维度隔离查询
    KEY `idx_company` (`company_id`, `platform_id`, `deleted_at`),
    -- SQL 内容变更检测 / 缓存
    KEY `idx_checksum` (`checksum`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'BI 模板 SQL 数据表';


INSERT INTO `bi_template` (`template_id`, `platform_id`, `code`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (1, 1, 'add', '', 1, '2026-01-15 09:28:45', '2026-01-15 10:01:13', NULL);
INSERT INTO `bi_template` (`template_id`, `platform_id`, `code`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (2, 1, 'del', '', 1, '2026-01-15 09:59:51', '2026-01-15 10:00:58', NULL);
INSERT INTO `bi_template` (`template_id`, `platform_id`, `code`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (3, 1, 'update', '', 1, '2026-01-15 10:00:04', '2026-01-15 10:01:00', NULL);
INSERT INTO `bi_template` (`template_id`, `platform_id`, `code`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (4, 1, 'select', '', 1, '2026-01-15 10:00:32', '2026-01-15 10:01:05', NULL);
INSERT INTO `bi_template` (`template_id`, `platform_id`, `code`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (5, 1, 'detail', '', 1, '2026-01-15 10:00:36', '2026-01-15 10:01:08', NULL);
INSERT INTO `bi_template` (`template_id`, `platform_id`, `code`, `name`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (6, 1, 'count', '', 1, '2026-01-15 11:03:38', '2026-01-15 11:03:38', NULL);

INSERT INTO `bi_template_data` (`td_id`, `platform_id`, `template_id`, `company_id`, `env`, `op_type`, `content`, `checksum`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (1, 1, 1, 0, 'test', 1, 'INSERT INTO `common_address` (`address_id`, `company_id`, `client_id`, `consignee`, `contact`, `phone`, `city_id`, `address`, `is_default`, `update_date`, `longitude`, `latitude`, `address_detail`, `district_id`)\nVALUES\n(null, 218908, 9932740, \'biAdd\', \'biAdd\', \'13096128259\', 110101, \'北京市 市辖区 东城区\', \'T\', \'2025-11-24 15:49:47\', 0, 0, \'sdaf\', NULL);', '', 1, '2026-01-15 09:30:34', '2026-01-15 11:07:19', NULL);
INSERT INTO `bi_template_data` (`td_id`, `platform_id`, `template_id`, `company_id`, `env`, `op_type`, `content`, `checksum`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (2, 1, 2, 0, 'test', 2, 'delete from common_address where company_id = 218908 and consignee = \'biAdd\';', '', 1, '2026-01-15 10:14:50', '2026-01-15 11:08:52', NULL);
INSERT INTO `bi_template_data` (`td_id`, `platform_id`, `template_id`, `company_id`, `env`, `op_type`, `content`, `checksum`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (3, 1, 3, 0, 'test', 3, 'UPDATE `common_address` SET `consignee` = \'zakiC.update.bi\' WHERE `address_id` = 4675682', '', 1, '2026-01-15 10:16:21', '2026-01-15 10:16:34', NULL);
INSERT INTO `bi_template_data` (`td_id`, `platform_id`, `template_id`, `company_id`, `env`, `op_type`, `content`, `checksum`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (4, 1, 4, 0, 'test', 401, 'select * from common_address where company_id = 218908', '', 1, '2026-01-15 10:16:32', '2026-01-15 11:00:09', NULL);
INSERT INTO `bi_template_data` (`td_id`, `platform_id`, `template_id`, `company_id`, `env`, `op_type`, `content`, `checksum`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (5, 1, 5, 0, 'test', 402, 'select * from common_address where company_id = 218908 and address_id = 4675640;', '', 1, '2026-01-15 11:01:15', '2026-01-15 11:01:22', NULL);
INSERT INTO `bi_template_data` (`td_id`, `platform_id`, `template_id`, `company_id`, `env`, `op_type`, `content`, `checksum`, `status`, `created_at`, `updated_at`, `deleted_at`) VALUES (6, 1, 6, 0, 'test', 403, 'select count(*) from common_address where company_id = 218908;', '', 1, '2026-01-15 11:01:32', '2026-01-15 11:02:05', NULL);