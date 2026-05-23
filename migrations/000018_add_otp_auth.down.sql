DROP TABLE IF EXISTS trusted_devices;
DROP TABLE IF EXISTS otp_codes;

ALTER TABLE users
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS phone_verified,
DROP COLUMN IF EXISTS phone_number;

