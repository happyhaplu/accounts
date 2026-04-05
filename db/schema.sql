-- ============================================================
--  Outcraftly Accounts — Identity Module
--  PostgreSQL Schema
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ------------------------------------------------------------
--  users
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id                  UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    email               VARCHAR(255)  UNIQUE NOT NULL,
    password_hash       TEXT          NOT NULL,
    reset_token         VARCHAR(255),
    reset_token_expires TIMESTAMPTZ,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email       ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users (reset_token);

-- ------------------------------------------------------------
--  Auto-update updated_at on every row change
-- ------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ------------------------------------------------------------
--  products.api_key  (migration — run once on existing DBs)
-- ------------------------------------------------------------
-- GORM AutoMigrate adds this column automatically on startup.
-- If you manage the schema manually, run:
--
--   ALTER TABLE products
--     ADD COLUMN IF NOT EXISTS api_key VARCHAR(120) UNIQUE;
--
-- Then back-fill existing rows:
--
--   UPDATE products
--   SET api_key = 'gour_ce_' || encode(gen_random_bytes(16), 'hex')
--   WHERE api_key IS NULL OR api_key = '';
--
