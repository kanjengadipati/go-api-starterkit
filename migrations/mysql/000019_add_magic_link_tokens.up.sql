CREATE TABLE IF NOT EXISTS magic_link_tokens (
    id CHAR(36) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    consumed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY idx_magic_link_tokens_hash (token_hash),
    INDEX idx_magic_link_tokens_user_id (user_id),
    INDEX idx_magic_link_tokens_lookup (token_hash, consumed_at, expires_at),
    CONSTRAINT fk_magic_link_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
