ALTER TABLE users
    DROP COLUMN IF EXISTS current_balance,
    DROP COLUMN IF EXISTS withdrawn_balance;

DROP TABLE IF EXISTS withdrawals;