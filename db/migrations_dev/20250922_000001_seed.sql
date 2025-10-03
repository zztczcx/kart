-- +goose Up
-- +goose StatementBegin
INSERT INTO products (id, name, category, price_cents) VALUES
  ('10','Chicken Waffle','Waffle',1299),
  ('11','Berry Waffle','Waffle',999),
  ('12','Latte','Beverage',499)
ON CONFLICT(id) DO NOTHING;

INSERT INTO coupons (code, presence_mask) VALUES
  ('HAPPYHRS', B'00000111'),  -- applies to 3 categories
  ('FIFTYOFF', B'00000011'),  -- applies to 2 categories
  ('SUPER100', B'00000001')   -- single category (will fail validation in service)
ON CONFLICT(code) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM order_items;
DELETE FROM orders;
DELETE FROM coupons;
DELETE FROM products;
-- +goose StatementEnd
