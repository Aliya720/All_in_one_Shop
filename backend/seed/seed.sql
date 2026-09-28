-- Idempotent sample data for local development / demos.
-- Safe to run multiple times thanks to ON CONFLICT DO NOTHING.

INSERT INTO categories (name, slug, description) VALUES
    ('Electronics', 'electronics', 'Phones, laptops, and gadgets'),
    ('Home & Kitchen', 'home-kitchen', 'Appliances and kitchenware'),
    ('Books', 'books', 'Fiction, non-fiction, and more'),
    ('Fashion', 'fashion', 'Clothing and accessories'),
    ('Sports & Outdoors', 'sports-outdoors', 'Gear for an active lifestyle')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO products (name, slug, description, price, stock, sku, category_id, image_url, is_active)
SELECT * FROM (VALUES
    ('Wireless Mouse', 'wireless-mouse', 'Ergonomic wireless mouse with USB receiver.', 799.00, 150, 'SKU-ELEC-001',
        (SELECT id FROM categories WHERE slug = 'electronics'), '', true),
    ('Mechanical Keyboard', 'mechanical-keyboard', 'RGB backlit mechanical keyboard, blue switches.', 3499.00, 60, 'SKU-ELEC-002',
        (SELECT id FROM categories WHERE slug = 'electronics'), '', true),
    ('27-inch Monitor', '27-inch-monitor', '1440p IPS monitor with 144Hz refresh rate.', 21999.00, 25, 'SKU-ELEC-003',
        (SELECT id FROM categories WHERE slug = 'electronics'), '', true),
    ('Stainless Steel Cookware Set', 'stainless-steel-cookware-set', '10-piece induction-compatible cookware set.', 5999.00, 40, 'SKU-HOME-001',
        (SELECT id FROM categories WHERE slug = 'home-kitchen'), '', true),
    ('Electric Kettle', 'electric-kettle', '1.7L rapid-boil electric kettle.', 1299.00, 80, 'SKU-HOME-002',
        (SELECT id FROM categories WHERE slug = 'home-kitchen'), '', true),
    ('The Pragmatic Programmer', 'the-pragmatic-programmer', 'Classic software craftsmanship book.', 899.00, 100, 'SKU-BOOK-001',
        (SELECT id FROM categories WHERE slug = 'books'), '', true),
    ('Atomic Habits', 'atomic-habits', 'Bestselling book on building good habits.', 599.00, 200, 'SKU-BOOK-002',
        (SELECT id FROM categories WHERE slug = 'books'), '', true),
    ('Cotton Crew-Neck T-Shirt', 'cotton-crew-neck-tshirt', 'Everyday soft cotton t-shirt, unisex fit.', 499.00, 300, 'SKU-FASH-001',
        (SELECT id FROM categories WHERE slug = 'fashion'), '', true),
    ('Running Shoes', 'running-shoes', 'Lightweight breathable running shoes.', 2999.00, 90, 'SKU-FASH-002',
        (SELECT id FROM categories WHERE slug = 'fashion'), '', true),
    ('Yoga Mat', 'yoga-mat', 'Non-slip 6mm thick yoga mat.', 899.00, 120, 'SKU-SPORT-001',
        (SELECT id FROM categories WHERE slug = 'sports-outdoors'), '', true),
    ('Adjustable Dumbbell Set', 'adjustable-dumbbell-set', 'Space-saving adjustable dumbbells, 2x20kg.', 8999.00, 15, 'SKU-SPORT-002',
        (SELECT id FROM categories WHERE slug = 'sports-outdoors'), '', true),
    ('Insulated Water Bottle', 'insulated-water-bottle', '1L stainless steel insulated bottle.', 799.00, 0, 'SKU-SPORT-003',
        (SELECT id FROM categories WHERE slug = 'sports-outdoors'), '', true)
) AS v(name, slug, description, price, stock, sku, category_id, image_url, is_active)
ON CONFLICT (slug) DO NOTHING;
