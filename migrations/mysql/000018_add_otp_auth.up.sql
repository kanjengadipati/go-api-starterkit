ALTER TABLE users
ADD COLUMN phone_number VARCHAR(20) UNIQUE,
ADD COLUMN phone_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

UPDATE users SET email_verified = is_verified WHERE email_verified = FALSE;

CREATE TABLE IF NOT EXISTS otp_codes (
    id CHAR(36) PRIMARY KEY,
    channel VARCHAR(50) NOT NULL,
    target VARCHAR(255) NOT NULL,
    code_hash TEXT NOT NULL,
    purpose VARCHAR(50) NOT NULL DEFAULT 'login',
    expires_at DATETIME NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    consumed BOOLEAN NOT NULL DEFAULT FALSE,
    provider VARCHAR(50),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_otp_codes_lookup (channel, target, purpose, consumed, created_at),
    INDEX idx_otp_codes_target_created (channel, target, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS trusted_devices (
    id CHAR(36) PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    device_hash VARCHAR(64) NOT NULL,
    device_name TEXT,
    last_used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY idx_trusted_devices_user_hash (user_id, device_hash),
    INDEX idx_trusted_devices_user_id (user_id),
    CONSTRAINT fk_trusted_devices_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
