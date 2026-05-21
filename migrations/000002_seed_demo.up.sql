-- Demo seed ma'lumotlar (faqat dev rejimi uchun)
-- Bu fayl docker-entrypoint orqali avtomatik bajariladi yoki qo'lda:
-- psql -d avtomakon -f migrations/000002_seed_demo.up.sql

-- Demo foydalanuvchilar
-- Parol = "demo1234" (Argon2id heshlangan — har bir foydalanuvchi uchun bir xil)
-- HASH faqat dev demo uchun. Production'da hech qachon shunday ishlatmang!
INSERT INTO users (id, phone, password_hash, full_name, username, role, is_business, is_verified, avatar_url, bio)
VALUES
    ('11111111-1111-1111-1111-111111111111',
     '+998900000001',
     '$argon2id$v=19$m=65536,t=3,p=2$ZGVtb2RlbW9kZW1vZGVtbw$placeholder_demo_hash_replace_me',
     'Jasur Bodywork', 'jasur_bodywork', 'master', TRUE, TRUE,
     'https://i.pravatar.cc/300?u=jasur', 'Avtomobil kuzov ishlari ustasi'),

    ('22222222-2222-2222-2222-222222222222',
     '+998900000002',
     '$argon2id$v=19$m=65536,t=3,p=2$ZGVtb2RlbW9kZW1vZGVtbw$placeholder_demo_hash_replace_me',
     'Akbar Ustaxona', 'akbar_ustaxona', 'master', TRUE, TRUE,
     'https://i.pravatar.cc/300?u=akbar', 'Dvigatel ta''mirlash'),

    ('33333333-3333-3333-3333-333333333333',
     '+998900000003',
     '$argon2id$v=19$m=65536,t=3,p=2$ZGVtb2RlbW9kZW1vZGVtbw$placeholder_demo_hash_replace_me',
     'Sardor Karimov', 'sardor', 'user', FALSE, FALSE,
     'https://i.pravatar.cc/300?u=sardor', 'Avtomobil ishqibozi 🚗')
ON CONFLICT (phone) DO NOTHING;

-- Demo postlar
INSERT INTO posts (id, author_id, caption, media_type, cover_url, location_name,
                   visibility, likes_count, comments_count, saves_count, shares_count)
VALUES
    ('a1111111-1111-1111-1111-111111111111',
     '11111111-1111-1111-1111-111111111111',
     'Mercedes-AMG GT 63 kuzov ishlari tugallandi. Oldingi va keyingi holat 🔥',
     'image',
     'https://images.unsplash.com/photo-1617531653332-bd46c24f2068?w=1080',
     'Toshkent', 'public', 8642, 345, 89, 23),

    ('a2222222-2222-2222-2222-222222222222',
     '22222222-2222-2222-2222-222222222222',
     'BMW M3 dvigatel diagnostikasi. Hammasi joyida ✅',
     'image',
     'https://images.unsplash.com/photo-1555215695-3004980ad54e?w=1080',
     'Toshkent', 'public', 1245, 56, 12, 5),

    ('a3333333-3333-3333-3333-333333333333',
     '11111111-1111-1111-1111-111111111111',
     'Audi RS6 to''liq detailing. Bularning yangi rangini ko''ring!',
     'image',
     'https://images.unsplash.com/photo-1606664515524-ed2f786a0bd6?w=1080',
     'Toshkent', 'public', 3421, 102, 45, 12)
ON CONFLICT (id) DO NOTHING;

-- Demo media
INSERT INTO post_media (post_id, url, type, width, height, order_index)
VALUES
    ('a1111111-1111-1111-1111-111111111111',
     'https://images.unsplash.com/photo-1617531653332-bd46c24f2068?w=1080',
     'image', 1080, 1920, 0),

    ('a2222222-2222-2222-2222-222222222222',
     'https://images.unsplash.com/photo-1555215695-3004980ad54e?w=1080',
     'image', 1080, 1920, 0),

    ('a3333333-3333-3333-3333-333333333333',
     'https://images.unsplash.com/photo-1606664515524-ed2f786a0bd6?w=1080',
     'image', 1080, 1920, 0)
ON CONFLICT DO NOTHING;

-- Hashtags
INSERT INTO hashtags (id, name, posts_count) VALUES
    ('b1111111-1111-1111-1111-111111111111', 'mercedes', 1),
    ('b2222222-2222-2222-2222-222222222222', 'amg', 1),
    ('b3333333-3333-3333-3333-333333333333', 'bodywork', 2),
    ('b4444444-4444-4444-4444-444444444444', 'bmw', 1),
    ('b5555555-5555-5555-5555-555555555555', 'audi', 1)
ON CONFLICT (name) DO NOTHING;

INSERT INTO post_hashtags (post_id, hashtag_id) VALUES
    ('a1111111-1111-1111-1111-111111111111', 'b1111111-1111-1111-1111-111111111111'),
    ('a1111111-1111-1111-1111-111111111111', 'b2222222-2222-2222-2222-222222222222'),
    ('a1111111-1111-1111-1111-111111111111', 'b3333333-3333-3333-3333-333333333333'),
    ('a2222222-2222-2222-2222-222222222222', 'b4444444-4444-4444-4444-444444444444'),
    ('a3333333-3333-3333-3333-333333333333', 'b5555555-5555-5555-5555-555555555555'),
    ('a3333333-3333-3333-3333-333333333333', 'b3333333-3333-3333-3333-333333333333')
ON CONFLICT DO NOTHING;
