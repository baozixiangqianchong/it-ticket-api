USE it_ticket;

CREATE TABLE IF NOT EXISTS invite_codes (
  id          BIGINT PRIMARY KEY AUTO_INCREMENT,
  code        VARCHAR(32)  NOT NULL,
  created_by  BIGINT       NOT NULL,
  expires_at  DATETIME     NOT NULL,
  used_at     DATETIME     NULL,
  used_by     BIGINT       NULL,
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_invite_code (code),
  KEY idx_invite_expires (expires_at, used_at),
  CONSTRAINT fk_invite_creator FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_invite_used_by FOREIGN KEY (used_by) REFERENCES users(id)
);
