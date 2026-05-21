-- post_likes ni post_reactions ga aylantirish (4 turdagi munosabat)
-- Mavjud like'lar avtomatik 'thumbs_up' bo'ladi

DO $$ BEGIN
    CREATE TYPE reaction_type AS ENUM ('thumbs_up', 'ok', 'handshake', 'thumbs_down');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Jadval nomini o'zgartirish
ALTER TABLE post_likes RENAME TO post_reactions;

-- Reaction turi kolonkasi
ALTER TABLE post_reactions
    ADD COLUMN IF NOT EXISTS reaction reaction_type NOT NULL DEFAULT 'thumbs_up';

-- Posts jadvalida reactions_count ham qo'shamiz (likes_count'ning nomi mantiqsiz endi)
ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS thumbs_up_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ok_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS handshake_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS thumbs_down_count INT NOT NULL DEFAULT 0;

-- Mavjud like'larni reactions_count'ga ko'chirish (har bir like = thumbs_up)
UPDATE posts SET thumbs_up_count = likes_count WHERE thumbs_up_count = 0;

-- Index yangilash
DROP INDEX IF EXISTS idx_post_likes_post;
CREATE INDEX IF NOT EXISTS idx_post_reactions_post ON post_reactions(post_id);
CREATE INDEX IF NOT EXISTS idx_post_reactions_user ON post_reactions(user_id);
