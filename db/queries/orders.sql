-- name: InsertOrder :exec
INSERT INTO orders (id, coupon_code)
VALUES ($1, $2);

-- name: InsertOrderItem :exec
INSERT INTO order_items (id, order_id, product_id, quantity)
VALUES ($1, $2, $3, $4);
