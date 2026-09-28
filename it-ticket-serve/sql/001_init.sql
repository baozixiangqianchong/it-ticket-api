CREATE DATABASE IF NOT EXISTS it_ticket
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE it_ticket;

CREATE TABLE IF NOT EXISTS users (
  id             BIGINT PRIMARY KEY AUTO_INCREMENT,
  email          VARCHAR(255) NOT NULL,
  password_hash  VARCHAR(255) NOT NULL,
  display_name   VARCHAR(64)  NOT NULL,
  role           ENUM('user','agent','admin') NOT NULL DEFAULT 'user',
  status         ENUM('active','disabled') NOT NULL DEFAULT 'active',
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_email (email)
);

CREATE TABLE IF NOT EXISTS tickets (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  title        VARCHAR(120) NOT NULL,
  description  TEXT         NOT NULL,
  category     ENUM('hardware','software','network','other') NOT NULL,
  priority     ENUM('p1','p2','p3') NOT NULL DEFAULT 'p2',
  status       ENUM('open','assigned','in_progress','pending','resolved','closed') NOT NULL DEFAULT 'open',
  creator_id   BIGINT NOT NULL,
  assignee_id  BIGINT NULL,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  closed_at    DATETIME NULL,
  CONSTRAINT fk_tickets_creator  FOREIGN KEY (creator_id)  REFERENCES users(id),
  CONSTRAINT fk_tickets_assignee FOREIGN KEY (assignee_id) REFERENCES users(id),
  KEY idx_tickets_status_updated   (status, updated_at),
  KEY idx_tickets_creator_updated  (creator_id, updated_at),
  KEY idx_tickets_assignee_updated (assignee_id, updated_at)
);

CREATE TABLE IF NOT EXISTS ticket_comments (
  id         BIGINT PRIMARY KEY AUTO_INCREMENT,
  ticket_id  BIGINT       NOT NULL,
  author_id  BIGINT       NOT NULL,
  body       VARCHAR(2000) NOT NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_comments_ticket FOREIGN KEY (ticket_id) REFERENCES tickets(id),
  CONSTRAINT fk_comments_author FOREIGN KEY (author_id) REFERENCES users(id),
  KEY idx_comments_ticket_created (ticket_id, created_at)
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  ticket_id   BIGINT      NOT NULL,
  actor_id    BIGINT      NOT NULL,
  action      VARCHAR(32) NOT NULL,
  from_status VARCHAR(32) NULL,
  to_status   VARCHAR(32) NOT NULL,
  reason      VARCHAR(500) NULL,
  created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_audit_ticket FOREIGN KEY (ticket_id) REFERENCES tickets(id),
  CONSTRAINT fk_audit_actor  FOREIGN KEY (actor_id)  REFERENCES users(id),
  KEY idx_audit_ticket_created (ticket_id, created_at)
);

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

CREATE TABLE IF NOT EXISTS account_audits (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  actor_id    BIGINT      NOT NULL,
  user_id     BIGINT      NOT NULL,
  action      VARCHAR(32) NOT NULL,
  from_value  VARCHAR(32) NULL,
  to_value    VARCHAR(32) NOT NULL,
  created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_account_audits_actor FOREIGN KEY (actor_id) REFERENCES users(id),
  CONSTRAINT fk_account_audits_user  FOREIGN KEY (user_id)  REFERENCES users(id),
  KEY idx_account_audits_created (created_at, id),
  KEY idx_account_audits_user    (user_id, created_at)
);
