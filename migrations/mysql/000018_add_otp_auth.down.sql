DROP TABLE IF EXISTS trusted_devices;
DROP TABLE IF EXISTS otp_codes;

ALTER TABLE users
DROP COLUMN email_verified,
DROP COLUMN phone_verified,
DROP COLUMN phone_number;

