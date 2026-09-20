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
  created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_email (email)
);

CREATE TABLE IF NOT EXISTS tickets (
  id           BIGINT PRIMARY KEY AUTO_INCREMENT,
  title        VARCHAR(120) NOT NULL,
  description  TEXT         NOT NULL,
  category     ENUM('hardware','software','network','other') NOT NULL,
  status       ENUM('open','assigned','in_progress','resolved','closed') NOT NULL DEFAULT 'open',
  creator_id   BIGINT NOT NULL,
  assignee_id  BIGINT NULL,
  created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
  created_at  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_audit_ticket FOREIGN KEY (ticket_id) REFERENCES tickets(id),
  CONSTRAINT fk_audit_actor  FOREIGN KEY (actor_id)  REFERENCES users(id),
  KEY idx_audit_ticket_created (ticket_id, created_at)
);
