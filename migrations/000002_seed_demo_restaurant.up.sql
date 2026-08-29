INSERT INTO restaurants (id, name) VALUES (1, 'Демо-суши')
ON CONFLICT (id) DO NOTHING;

SELECT setval('restaurants_id_seq', GREATEST((SELECT MAX(id) FROM restaurants), 1));

INSERT INTO menu_items (restaurant_id, name, price_cents, stock) VALUES
    (1, 'Ролл Калифорния', 45000, 10),
    (1, 'Ролл Филадельфия', 52000, 10)
ON CONFLICT DO NOTHING;
