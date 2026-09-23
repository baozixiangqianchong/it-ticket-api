-- V2 P0：给已有表加字段，不建新表。
-- tickets.closed_at：关闭 / 撤回时间，用来算 7 天重开窗口
-- audit_logs.reason：重开、管理员代关、撤回时写下的原因
-- 新库直接跑 001_init.sql 即可（里面已经有这两列）。
-- 本文件可重复执行：列已存在就跳过。

USE it_ticket;

SET @exist := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = 'it_ticket' AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'closed_at'
);
SET @sql := IF(@exist = 0,
  'ALTER TABLE tickets ADD COLUMN closed_at DATETIME NULL AFTER updated_at',
  'SELECT ''tickets.closed_at 已存在'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = 'it_ticket' AND TABLE_NAME = 'audit_logs' AND COLUMN_NAME = 'reason'
);
SET @sql := IF(@exist = 0,
  'ALTER TABLE audit_logs ADD COLUMN reason VARCHAR(500) NULL AFTER to_status',
  'SELECT ''audit_logs.reason 已存在'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
