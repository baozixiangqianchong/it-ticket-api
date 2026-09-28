-- 账号变更审计：改角色、停用 / 启用。已有库执行本文件。
USE it_ticket;

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
