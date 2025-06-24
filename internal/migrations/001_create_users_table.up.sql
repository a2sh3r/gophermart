CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       login TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMP DEFAULT NOW()
);