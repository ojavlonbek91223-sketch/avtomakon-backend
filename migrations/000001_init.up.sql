-- AvtoMakon dastlabki sxema
-- Yaratilgan: 2026-05-17

-- ============================================================================
-- KENGAYTMALAR
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- ============================================================================
-- ENUM TIPLAR
-- ============================================================================
CREATE TYPE user_role AS ENUM ('user', 'master', 'seller', 'admin');
CREATE TYPE business_type AS ENUM ('master', 'seller');
CREATE TYPE application_status AS ENUM ('pending', 'approved', 'rejected', 'requires_changes');
CREATE TYPE post_media_type AS ENUM ('image', 'video', 'carousel');
CREATE TYPE post_visibility AS ENUM ('public', 'followers', 'private');
CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'paid', 'shipped', 'delivered', 'cancelled', 'refunded');
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded');
CREATE TYPE payment_method AS ENUM ('card', 'click', 'payme', 'cash');
CREATE TYPE delivery_method AS ENUM ('pickup', 'courier', 'post');
CREATE TYPE conversation_type AS ENUM ('direct', 'group');
CREATE TYPE message_type AS ENUM ('text', 'image', 'file', 'location', 'product', 'system');
CREATE TYPE notification_type AS ENUM ('like', 'comment', 'follow', 'message', 'order_update', 'business_approved', 'review');
CREATE TYPE otp_purpose AS ENUM ('signup', 'login', 'reset_password', 'verify_phone');

-- ============================================================================
-- FOYDALANUVCHILAR
-- ============================================================================
CREATE TABLE users (
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

CREATE INDEX idx_users_phone ON users(phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE INDEX idx_users_role ON users(role);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address  INET,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);

CREATE TABLE otp_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL,
    code_hash  VARCHAR(255) NOT NULL,
    purpose    otp_purpose NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_identifier ON otp_codes(identifier, purpose) WHERE used_at IS NULL;

-- ============================================================================
-- BIZNES
-- ============================================================================
CREATE TABLE business_applications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type              business_type NOT NULL,
    business_name     VARCHAR(150) NOT NULL,
    contact_phone     VARCHAR(20) NOT NULL,
    address           TEXT NOT NULL,
    location          GEOGRAPHY(POINT, 4326),
    experience_years  INTEGER,
    description       TEXT,
    workplace_photos  TEXT[],
    document_url      TEXT,
    status            application_status NOT NULL DEFAULT 'pending',
    admin_notes       TEXT,
    reviewed_by       UUID REFERENCES users(id),
    reviewed_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_applications_user ON business_applications(user_id);
CREATE INDEX idx_applications_status ON business_applications(status);

CREATE TABLE businesses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES users(id),
    application_id  UUID REFERENCES business_applications(id),
    type            business_type NOT NULL,
    name            VARCHAR(150) NOT NULL,
    slug            VARCHAR(150) UNIQUE NOT NULL,
    description     TEXT,
    phone           VARCHAR(20) NOT NULL,
    address         TEXT NOT NULL,
    location        GEOGRAPHY(POINT, 4326) NOT NULL,
    working_hours   JSONB,
    cover_image_url TEXT,
    gallery         TEXT[],
    rating_avg      DECIMAL(2, 1) NOT NULL DEFAULT 0,
    rating_count    INTEGER NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_businesses_location ON businesses USING GIST(location);
CREATE INDEX idx_businesses_type ON businesses(type) WHERE is_active = TRUE;
CREATE INDEX idx_businesses_owner ON businesses(owner_id);
CREATE INDEX idx_businesses_name_trgm ON businesses USING GIN(name gin_trgm_ops);

CREATE TABLE business_services (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id      UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name             VARCHAR(150) NOT NULL,
    description      TEXT,
    price_from       DECIMAL(12, 2),
    price_to         DECIMAL(12, 2),
    currency         VARCHAR(3) NOT NULL DEFAULT 'UZS',
    duration_minutes INTEGER,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_services_business ON business_services(business_id);

-- ============================================================================
-- IJTIMOIY (FEED)
-- ============================================================================
CREATE TABLE posts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    caption        TEXT,
    media_type     post_media_type NOT NULL,
    cover_url      TEXT,
    location_name  VARCHAR(200),
    location       GEOGRAPHY(POINT, 4326),
    visibility     post_visibility NOT NULL DEFAULT 'public',
    is_published   BOOLEAN NOT NULL DEFAULT TRUE,
    likes_count    INTEGER NOT NULL DEFAULT 0,
    comments_count INTEGER NOT NULL DEFAULT 0,
    shares_count   INTEGER NOT NULL DEFAULT 0,
    saves_count    INTEGER NOT NULL DEFAULT 0,
    views_count    BIGINT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_posts_author ON posts(author_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_feed ON posts(created_at DESC)
    WHERE deleted_at IS NULL AND is_published = TRUE AND visibility = 'public';

CREATE TABLE post_media (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id          UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    url              TEXT NOT NULL,
    thumbnail_url    TEXT,
    type             VARCHAR(10) NOT NULL CHECK (type IN ('image', 'video')),
    duration_seconds INTEGER,
    width            INTEGER,
    height           INTEGER,
    order_index      INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_post_media_post ON post_media(post_id, order_index);

CREATE TABLE post_likes (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE INDEX idx_post_likes_post ON post_likes(post_id);

CREATE TABLE post_saves (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES comments(id) ON DELETE CASCADE,
    text        TEXT NOT NULL,
    likes_count INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_comments_post ON comments(post_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_parent ON comments(parent_id) WHERE parent_id IS NOT NULL;

CREATE TABLE comment_likes (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id)
);

CREATE TABLE hashtags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    posts_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE post_hashtags (
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    hashtag_id UUID NOT NULL REFERENCES hashtags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, hashtag_id)
);

CREATE INDEX idx_post_hashtags_hashtag ON post_hashtags(hashtag_id);

CREATE TABLE follows (
    follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id <> following_id)
);

CREATE INDEX idx_follows_following ON follows(following_id);

-- ============================================================================
-- MARKET
-- ============================================================================
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id   UUID REFERENCES categories(id),
    name        JSONB NOT NULL,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    icon_url    TEXT,
    order_index INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_categories_parent ON categories(parent_id);

CREATE TABLE products (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id        UUID NOT NULL REFERENCES businesses(id),
    category_id      UUID NOT NULL REFERENCES categories(id),
    name             VARCHAR(200) NOT NULL,
    slug             VARCHAR(250) UNIQUE NOT NULL,
    description      TEXT,
    brand            VARCHAR(100),
    sku              VARCHAR(100),
    price            DECIMAL(12, 2) NOT NULL,
    original_price   DECIMAL(12, 2),
    currency         VARCHAR(3) NOT NULL DEFAULT 'UZS',
    stock_quantity   INTEGER NOT NULL DEFAULT 0,
    min_order_qty    INTEGER NOT NULL DEFAULT 1,
    rating_avg       DECIMAL(2, 1) NOT NULL DEFAULT 0,
    rating_count     INTEGER NOT NULL DEFAULT 0,
    sales_count      INTEGER NOT NULL DEFAULT 0,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category ON products(category_id) WHERE is_active = TRUE;
CREATE INDEX idx_products_seller ON products(seller_id);
CREATE INDEX idx_products_featured ON products(is_featured) WHERE is_active = TRUE AND is_featured = TRUE;
CREATE INDEX idx_products_name_trgm ON products USING GIN(name gin_trgm_ops);

CREATE TABLE product_images (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    order_index INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_product_images_product ON product_images(product_id, order_index);

CREATE TABLE product_attributes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    key        VARCHAR(50) NOT NULL,
    value      VARCHAR(150) NOT NULL
);

CREATE INDEX idx_product_attrs_product ON product_attributes(product_id);

CREATE TABLE promotions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(150) NOT NULL,
    subtitle    VARCHAR(250),
    image_url   TEXT,
    link_type   VARCHAR(20) NOT NULL CHECK (link_type IN ('category', 'product', 'external_url', 'none')),
    link_target VARCHAR(250),
    starts_at   TIMESTAMPTZ NOT NULL,
    ends_at     TIMESTAMPTZ NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    order_index INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_promotions_active ON promotions(is_active, starts_at, ends_at);

CREATE TABLE carts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cart_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id    UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INTEGER NOT NULL CHECK (quantity > 0),
    added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cart_id, product_id)
);

CREATE TABLE orders (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number           VARCHAR(20) UNIQUE NOT NULL,
    user_id                UUID NOT NULL REFERENCES users(id),
    status                 order_status NOT NULL DEFAULT 'pending',
    subtotal               DECIMAL(12, 2) NOT NULL,
    delivery_fee           DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total                  DECIMAL(12, 2) NOT NULL,
    currency               VARCHAR(3) NOT NULL DEFAULT 'UZS',
    delivery_address       JSONB NOT NULL,
    delivery_method        delivery_method NOT NULL,
    payment_method         payment_method NOT NULL,
    payment_status         payment_status NOT NULL DEFAULT 'pending',
    payment_transaction_id VARCHAR(100),
    note                   TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_status ON orders(status);

CREATE TABLE order_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id          UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id        UUID NOT NULL REFERENCES products(id),
    seller_id         UUID NOT NULL REFERENCES businesses(id),
    quantity          INTEGER NOT NULL CHECK (quantity > 0),
    price_at_purchase DECIMAL(12, 2) NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_seller ON order_items(seller_id);

-- ============================================================================
-- SHARHLAR
-- ============================================================================
CREATE TABLE reviews (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type        VARCHAR(20) NOT NULL CHECK (target_type IN ('business', 'product', 'order')),
    target_id          UUID NOT NULL,
    order_id           UUID REFERENCES orders(id),
    rating             SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    text               TEXT,
    images             TEXT[],
    seller_reply       TEXT,
    seller_replied_at  TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (author_id, target_type, target_id)
);

CREATE INDEX idx_reviews_target ON reviews(target_type, target_id, created_at DESC);
CREATE INDEX idx_reviews_author ON reviews(author_id);

-- ============================================================================
-- CHAT
-- ============================================================================
CREATE TABLE conversations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type             conversation_type NOT NULL DEFAULT 'direct',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_message_id  UUID,
    last_message_at  TIMESTAMPTZ
);

CREATE INDEX idx_conversations_last_message ON conversations(last_message_at DESC NULLS LAST);

CREATE TABLE conversation_members (
    conversation_id      UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_read_message_id UUID,
    is_muted             BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_conv_members_user ON conversation_members(user_id);

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES users(id),
    type            message_type NOT NULL DEFAULT 'text',
    text            TEXT,
    media_url       TEXT,
    metadata        JSONB,
    reply_to_id     UUID REFERENCES messages(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at       TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_messages_conv ON messages(conversation_id, created_at DESC) WHERE deleted_at IS NULL;

-- conversations.last_message_id ga deferred FK
ALTER TABLE conversations
    ADD CONSTRAINT fk_conv_last_message
    FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE SET NULL
    DEFERRABLE INITIALLY DEFERRED;

-- ============================================================================
-- BILDIRISHNOMALAR
-- ============================================================================
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        notification_type NOT NULL,
    actor_id    UUID REFERENCES users(id),
    entity_type VARCHAR(30),
    entity_id   UUID,
    title       VARCHAR(200) NOT NULL,
    body        TEXT,
    data        JSONB,
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_unread ON notifications(user_id, created_at DESC) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_user_all ON notifications(user_id, created_at DESC);

CREATE TABLE push_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT UNIQUE NOT NULL,
    platform   VARCHAR(10) NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_tokens_user ON push_tokens(user_id);

-- ============================================================================
-- YORDAMCHI
-- ============================================================================
CREATE TABLE uploaded_files (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose    VARCHAR(50) NOT NULL,
    filename   VARCHAR(255) NOT NULL,
    mime_type  VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    url        TEXT NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_owner ON uploaded_files(owner_id);

CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID REFERENCES users(id),
    action      VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id   UUID,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_action ON audit_logs(action, created_at DESC);

CREATE TABLE reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id),
    target_type VARCHAR(20) NOT NULL CHECK (target_type IN ('post', 'user', 'product', 'comment', 'business')),
    target_id   UUID NOT NULL,
    reason      VARCHAR(50) NOT NULL,
    description TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'dismissed', 'action_taken')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reports_target ON reports(target_type, target_id);
CREATE INDEX idx_reports_status ON reports(status);

-- ============================================================================
-- TRIGGERLAR — updated_at avtomatik yangilash
-- ============================================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_businesses_updated_at BEFORE UPDATE ON businesses
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_carts_updated_at BEFORE UPDATE ON carts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
