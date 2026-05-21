ALTER TABLE post_reactions DROP COLUMN IF EXISTS reaction;
ALTER TABLE post_reactions RENAME TO post_likes;
ALTER TABLE posts DROP COLUMN IF EXISTS thumbs_up_count;
ALTER TABLE posts DROP COLUMN IF EXISTS ok_count;
ALTER TABLE posts DROP COLUMN IF EXISTS handshake_count;
ALTER TABLE posts DROP COLUMN IF EXISTS thumbs_down_count;
DROP TYPE IF EXISTS reaction_type;
