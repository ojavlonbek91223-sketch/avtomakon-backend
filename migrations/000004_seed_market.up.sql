-- Demo Market: kategoriyalar, mahsulotlar, aksiyalar

-- Kategoriyalar
INSERT INTO categories (id, parent_id, name, slug, order_index) VALUES
    ('d1111111-1111-1111-1111-111111111111', NULL,
     '{"uz":"Dvigatel","ru":"Двигатель","en":"Engine"}'::jsonb,
     'dvigatel', 1),
    ('d2222222-2222-2222-2222-222222222222', NULL,
     '{"uz":"Shinalar","ru":"Шины","en":"Tires"}'::jsonb,
     'shinalar', 2),
    ('d3333333-3333-3333-3333-333333333333', NULL,
     '{"uz":"Kuzov qismi","ru":"Кузовные детали","en":"Body parts"}'::jsonb,
     'kuzov-qismi', 3),
    ('d4444444-4444-4444-4444-444444444444', NULL,
     '{"uz":"Moylar","ru":"Масла","en":"Oils"}'::jsonb,
     'moylar', 4),
    ('d5555555-5555-5555-5555-555555555555', NULL,
     '{"uz":"Aksessuarlar","ru":"Аксессуары","en":"Accessories"}'::jsonb,
     'aksessuarlar', 5)
ON CONFLICT (slug) DO NOTHING;

-- Mahsulotlar (sotuvchi: Shina Market va AutoParts Tashkent)
INSERT INTO products (id, seller_id, category_id, name, slug, description, brand, sku,
                      price, original_price, currency, stock_quantity,
                      rating_avg, rating_count, sales_count, is_featured, is_active)
VALUES
    ('e1111111-1111-1111-1111-111111111111',
     'c5555555-5555-5555-5555-555555555555',
     'd4444444-4444-4444-4444-444444444444',
     'Mobil 1 5W-30 sintetik moy (4L)',
     'mobil-1-5w-30-4l',
     'Original Mobil 1 sintetik motor moyi. Yuqori sifatli, barcha mavsumlar uchun.',
     'Mobil 1', 'MOB-5W30-4L',
     285000, 350000, 'UZS', 24,
     4.8, 156, 89, TRUE, TRUE),

    ('e2222222-2222-2222-2222-222222222222',
     'c4444444-4444-4444-4444-444444444444',
     'd2222222-2222-2222-2222-222222222222',
     'Michelin Pilot Sport 4 225/45 R17',
     'michelin-pilot-sport-4-225-45-r17',
     'Yozgi yuqori tezlikli shinalar. Premium sifat.',
     'Michelin', 'MIC-PS4-225-45-17',
     890000, 1050000, 'UZS', 12,
     4.9, 89, 45, TRUE, TRUE),

    ('e3333333-3333-3333-3333-333333333333',
     'c5555555-5555-5555-5555-555555555555',
     'd1111111-1111-1111-1111-111111111111',
     'NGK Iridium IX svechalar (4 dona)',
     'ngk-iridium-ix-4pcs',
     'Premium iridium uchqun svechalari. 4 ta to''plam.',
     'NGK', 'NGK-IX-4',
     320000, 380000, 'UZS', 35,
     4.7, 67, 124, FALSE, TRUE),

    ('e4444444-4444-4444-4444-444444444444',
     'c4444444-4444-4444-4444-444444444444',
     'd2222222-2222-2222-2222-222222222222',
     'Continental ContiPremiumContact 195/65 R15',
     'continental-cpc-195-65-r15',
     'Yozgi shinalar, premium komfort va xavfsizlik.',
     'Continental', 'CON-CPC-195-65-15',
     620000, NULL, 'UZS', 28,
     4.6, 102, 78, FALSE, TRUE),

    ('e5555555-5555-5555-5555-555555555555',
     'c5555555-5555-5555-5555-555555555555',
     'd4444444-4444-4444-4444-444444444444',
     'Castrol Edge 5W-40 (5L)',
     'castrol-edge-5w-40-5l',
     'Full sintetik dvigatel moyi, FST texnologiyasi.',
     'Castrol', 'CAS-EDGE-5L',
     410000, 480000, 'UZS', 18,
     4.7, 134, 67, TRUE, TRUE),

    ('e6666666-6666-6666-6666-666666666666',
     'c5555555-5555-5555-5555-555555555555',
     'd1111111-1111-1111-1111-111111111111',
     'Bosch havo filtri (universal)',
     'bosch-air-filter-universal',
     'Yuqori filtratsiya darajasi, asl Bosch.',
     'Bosch', 'BOSCH-AF-U',
     145000, 180000, 'UZS', 56,
     4.5, 78, 234, FALSE, TRUE)
ON CONFLICT (slug) DO NOTHING;

-- Mahsulot rasmlari
INSERT INTO product_images (product_id, url, order_index) VALUES
    ('e1111111-1111-1111-1111-111111111111',
     'https://images.unsplash.com/photo-1635764928275-cb1ba47b3f30?w=600', 0),
    ('e2222222-2222-2222-2222-222222222222',
     'https://images.unsplash.com/photo-1542362567-b07e54358753?w=600', 0),
    ('e3333333-3333-3333-3333-333333333333',
     'https://images.unsplash.com/photo-1632823469850-1b7b1e8b7ecc?w=600', 0),
    ('e4444444-4444-4444-4444-444444444444',
     'https://images.unsplash.com/photo-1580273916550-e323be2ae537?w=600', 0),
    ('e5555555-5555-5555-5555-555555555555',
     'https://images.unsplash.com/photo-1486754735734-325b5831c3ad?w=600', 0),
    ('e6666666-6666-6666-6666-666666666666',
     'https://images.unsplash.com/photo-1492144534655-ae79c964c9d7?w=600', 0)
ON CONFLICT DO NOTHING;

-- Aksiyalar
INSERT INTO promotions (id, title, subtitle, image_url, link_type, link_target,
                        starts_at, ends_at, is_active, order_index)
VALUES
    ('f1111111-1111-1111-1111-111111111111',
     'BAHORGI CHEGIRMALAR',
     'Barcha moy va filtrlarga 30% chegirma',
     'https://images.unsplash.com/photo-1486754735734-325b5831c3ad?w=1200',
     'category', 'moylar',
     NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days', TRUE, 1),

    ('f2222222-2222-2222-2222-222222222222',
     'Premium shinalar — yozga tayyor',
     'Michelin va Continental shinalariga 15% chegirma',
     'https://images.unsplash.com/photo-1542362567-b07e54358753?w=1200',
     'category', 'shinalar',
     NOW() - INTERVAL '1 day', NOW() + INTERVAL '20 days', TRUE, 2),

    ('f3333333-3333-3333-3333-333333333333',
     'Yangi mahsulotlar',
     'Eng yangi avto aksessuarlar',
     'https://images.unsplash.com/photo-1492144534655-ae79c964c9d7?w=1200',
     'category', 'aksessuarlar',
     NOW() - INTERVAL '1 day', NOW() + INTERVAL '60 days', TRUE, 3)
ON CONFLICT (id) DO NOTHING;
