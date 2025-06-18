ALTER TABLE users
    ADD COLUMN IF NOT EXISTS current_balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS withdrawn_balance NUMERIC(12,2) NOT NULL DEFAULT 0;

ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'NEW';

CREATE TABLE IF NOT EXISTS withdrawals (
                                           id SERIAL PRIMARY KEY,
                                           order_number VARCHAR(255) NOT NULL,
                                           sum NUMERIC(12,2) NOT NULL CHECK (sum > 0),
                                           processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
                                           user_id BIGINT NOT NULL REFERENCES users(id)
);