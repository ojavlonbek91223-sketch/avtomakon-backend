-- Test foydalanuvchi: telefon +998887360806, parol "Javlonbek-03"
-- Hesh `cmd/hashtool` orqali yaratilgan (Argon2id, OWASP 2024 parametrlari)

INSERT INTO users (
    id, phone, password_hash, full_name, username, role,
    is_verified, is_business, language, phone_verified_at
)
VALUES (
    '00000000-1111-2222-3333-444444444444',
    '+998887360806',
    '$argon2id$v=19$m=65536,t=3,p=2$6/nm/tH4mJ+0UmYDyruiPg$bSjjwH9TLBL4jSdQixssSmzFJ4LVUcpP+STeC1zavt4',
    'Javlonbek',
    'javlonbek',
    'user',
    TRUE,
    FALSE,
    'uz',
    NOW()
)
ON CONFLICT (phone) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    is_verified = TRUE,
    phone_verified_at = NOW();
