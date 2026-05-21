-- Demo suhbatlar (Sardor <-> Jasur, Sardor <-> Akbar)

-- Suhbatlar
INSERT INTO conversations (id, type, created_at, last_message_at) VALUES
    ('aa111111-1111-1111-1111-111111111111', 'direct',
     NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 hours'),
    ('aa222222-2222-2222-2222-222222222222', 'direct',
     NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- A'zolar
INSERT INTO conversation_members (conversation_id, user_id) VALUES
    ('aa111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333'),
    ('aa111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111'),
    ('aa222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333'),
    ('aa222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222')
ON CONFLICT DO NOTHING;

-- Xabarlar (Sardor <-> Jasur)
INSERT INTO messages (id, conversation_id, sender_id, type, text, created_at) VALUES
    ('bb111111-1111-1111-1111-111111111111',
     'aa111111-1111-1111-1111-111111111111',
     '33333333-3333-3333-3333-333333333333',
     'text', 'Salom Jasur aka, mashinamga bodywork ishlari kerak edi',
     NOW() - INTERVAL '5 days'),
    ('bb222222-2222-2222-2222-222222222222',
     'aa111111-1111-1111-1111-111111111111',
     '11111111-1111-1111-1111-111111111111',
     'text', 'Salom Sardor! Albatta, qachon olib kelolasiz?',
     NOW() - INTERVAL '4 days'),
    ('bb333333-3333-3333-3333-333333333333',
     'aa111111-1111-1111-1111-111111111111',
     '11111111-1111-1111-1111-111111111111',
     'text', 'Kuzov ishlarini 3 kunda tugatamiz',
     NOW() - INTERVAL '2 hours')
ON CONFLICT (id) DO NOTHING;

-- Xabarlar (Sardor <-> Akbar)
INSERT INTO messages (id, conversation_id, sender_id, type, text, created_at) VALUES
    ('bb444444-4444-4444-4444-444444444444',
     'aa222222-2222-2222-2222-222222222222',
     '33333333-3333-3333-3333-333333333333',
     'text', 'Akbar aka, dvigatel diagnostika qilish kerak edi',
     NOW() - INTERVAL '3 days'),
    ('bb555555-5555-5555-5555-555555555555',
     'aa222222-2222-2222-2222-222222222222',
     '22222222-2222-2222-2222-222222222222',
     'text', 'Ha, dvigatel ta''mirini qilsak bo''ladi. Keling ertaga ko''ramiz',
     NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- Conversations'ning last_message_id ni yangilaymiz
UPDATE conversations SET last_message_id = 'bb333333-3333-3333-3333-333333333333'
    WHERE id = 'aa111111-1111-1111-1111-111111111111';
UPDATE conversations SET last_message_id = 'bb555555-5555-5555-5555-555555555555'
    WHERE id = 'aa222222-2222-2222-2222-222222222222';
