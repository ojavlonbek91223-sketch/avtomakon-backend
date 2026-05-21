-- Uzun videolar (faqat usta va sotuvchilar yuborishi mumkin)
-- YouTube'dan farqli — avto sohasiga moslangan

CREATE TABLE IF NOT EXISTS long_videos (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    video_url       TEXT NOT NULL,
    thumbnail_url   TEXT,
    duration_sec    INT NOT NULL DEFAULT 0,
    category        VARCHAR(50),
    views_count     BIGINT NOT NULL DEFAULT 0,
    reactions_count INT NOT NULL DEFAULT 0,
    comments_count  INT NOT NULL DEFAULT 0,
    is_published    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_long_videos_author ON long_videos(author_id);
CREATE INDEX IF NOT EXISTS idx_long_videos_created ON long_videos(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_long_videos_category ON long_videos(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_long_videos_title_trgm ON long_videos USING GIN(title gin_trgm_ops);

-- Trigger updated_at uchun
DROP TRIGGER IF EXISTS trg_long_videos_updated_at ON long_videos;
CREATE TRIGGER trg_long_videos_updated_at BEFORE UPDATE ON long_videos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Seed demo videolar (master/seller userlardan)
INSERT INTO long_videos (id, author_id, title, description, video_url, thumbnail_url, duration_sec, category, views_count, reactions_count, comments_count)
VALUES
    ('v1111111-1111-1111-1111-111111111111',
     '11111111-1111-1111-1111-111111111111',
     'Mercedes-AMG GT 63 — kuzov bo''yash jarayoni (to''liq)',
     'Avtomobil kuzovini noldan bo''yash. Materiallar, asboblar, jarayon. 5 yillik tajriba bilan ulashaman.',
     'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4',
     'https://images.unsplash.com/photo-1617531653332-bd46c24f2068?w=800',
     1820, 'bo''yash', 12450, 892, 67),

    ('v2222222-2222-2222-2222-222222222222',
     '22222222-2222-2222-2222-222222222222',
     'BMW M3 dvigatel diagnostikasi — boshlovchilar uchun',
     'Dvigatel bilan bog''liq asosiy nosozliklarni qanday aniqlash. Diagnostika skanerlari, OBD2.',
     'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4',
     'https://images.unsplash.com/photo-1555215695-3004980ad54e?w=800',
     2340, 'dvigatel', 8920, 543, 32),

    ('v3333333-3333-3333-3333-333333333333',
     '11111111-1111-1111-1111-111111111111',
     'Audi RS6 — detailing 0 dan tugaguncha',
     'Premium detailing. Ceramic coating, paint correction, interior detailing.',
     'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4',
     'https://images.unsplash.com/photo-1606664515524-ed2f786a0bd6?w=800',
     2780, 'detailing', 15670, 1203, 89)
ON CONFLICT (id) DO NOTHING;
