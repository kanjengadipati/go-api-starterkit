ALTER TABLE users
ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20) UNIQUE,
ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT FALSE;

UPDATE users SET email_verified = is_verified WHERE email_verified = FALSE;

CREATE TABLE IF NOT EXISTS otp_codes (
    id UUID PRIMARY KEY,
    channel VARCHAR(50) NOT NULL,
    target VARCHAR(255) NOT NULL,
    code_hash TEXT NOT NULL,
    purpose VARCHAR(50) NOT NULL DEFAULT 'login',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    consumed BOOLEAN NOT NULL DEFAULT FALSE,
    provider VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_otp_codes_lookup ON otp_codes (channel, target, purpose, consumed, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_otp_codes_target_created ON otp_codes (channel, target, created_at DESC);

CREATE TABLE IF NOT EXISTS trusted_devices (
    id UUID PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_hash TEXT NOT NULL,
    device_name TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, device_hash)
);

CREATE INDEX IF NOT EXISTS idx_trusted_devices_user_id ON trusted_devices (user_id);

