CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    role_id INTEGER,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    password_updated_at TIMESTAMP
);