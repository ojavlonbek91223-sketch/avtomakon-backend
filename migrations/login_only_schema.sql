-- Faqat login uchun zarur jadvallar (PostGIS'siz).
-- Login ishlatish uchun: users, refresh_tokens, otp_codes
-- Boshqa jadvallar PostGIS o'rnatilgach qo'shiladi.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ENUM tip
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'master', 'seller', 'admin');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE otp_purpose AS ENUM ('signup', 'login', 'reset_password', 'verify_phone');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Users
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) UNIQUE,
    phone           VARCHAR(20) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    full_name       VARCHAR(100) NOT NULL,
    username        VARCHAR(50) UNIQUE,
    avatar_url      TEXT,
    bio             VARCHAR(200),
    role            user_role NOT NULL DEFAULT 'user',
    is_business     BOOLEAN NOT NULL DEFAULT FALSE,
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    language        VARCHAR(5) NOT NULL DEFAULT 'uz',
    country_code    VARCHAR(2),
    last_lat        DECIMAL(10, 7),
    last_lng        DECIMAL(10, 7),
    last_active_at  TIMESTAMPTZ,
    posts_count     INTEGER NOT NULL DEFAULT 0,
    followers_count INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL;

-- Refresh tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address  INET,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- OTP codes
CREATE TABLE IF NOT EXISTS otp_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL,
    code_hash  VARCHAR(255) NOT NULL,
    purpose    otp_purpose NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- updated_at trigger
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Test foydalanuvchi
INSERT INTO users (
    id, phone, password_hash, full_name, username, role,
    is_verified, language, phone_verified_at
)
VALUES (
    '00000000-1111-2222-3333-444444444444',
    '+998887360806',
    '$argon2id$v=19$m=65536,t=3,p=2$6/nm/tH4mJ+0UmYDyruiPg$bSjjwH9TLBL4jSdQixssSmzFJ4LVUcpP+STeC1zavt4',
    'Javlonbek',
    'javlonbek',
    'user',
    TRUE,
    'uz',
    NOW()
)
ON CONFLICT (phone) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    is_verified = TRUE,
    phone_verified_at = NOW();
