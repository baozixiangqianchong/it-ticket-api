-- V2 P1：优先级、等用户、停用账号、站内通知。不做附件。
-- 新库直接跑更新后的 001_init.sql 即可。
-- 本文件可重复执行。

USE it_ticket;

SET @exist := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = 'it_ticket' AND TABLE_NAME = 'users' AND COLUMN_NAME = 'status'
);
SET @sql := IF(@exist = 0,
  'ALTER TABLE users ADD COLUMN status ENUM(''active'',''disabled'') NOT NULL DEFAULT ''active'' AFTER role',
  'SELECT ''users.status 已存在'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exist := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = 'it_ticket' AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'priority'
);
SET @sql := IF(@exist = 0,
  'ALTER TABLE tickets ADD COLUMN priority ENUM(''p1'',''p2'',''p3'') NOT NULL DEFAULT ''p2'' AFTER category',
  'SELECT ''tickets.priority 已存在'' AS info'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE tickets
  MODIFY COLUMN status ENUM('open','assigned','in_progress','pending','resolved','closed') NOT NULL DEFAULT 'open';

CREATE TABLE IF NOT EXISTS notifications (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id    BIGINT       NOT NULL,
  ticket_id  BIGINT       NULL,
  type       VARCHAR(32)  NOT NULL,
  title      VARCHAR(120) NOT NULL,
  body       VARCHAR(500) NULL,
  read_at    DATETIME     NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_notif_user   FOREIGN KEY (user_id)   REFERENCES users(id),
  CONSTRAINT fk_notif_ticket FOREIGN KEY (ticket_id) REFERENCES tickets(id),
  KEY idx_notif_user_created (user_id, created_at),
  KEY idx_notif_user_unread  (user_id, read_at)
);
