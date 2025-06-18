CREATE TABLE IF NOT EXISTS orders (
                                      number TEXT PRIMARY KEY,
                                      status TEXT NOT NULL,
                                      accrual DOUBLE PRECISION,
                                      uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                      user_id BIGINT NOT NULL REFERENCES users(id)
);