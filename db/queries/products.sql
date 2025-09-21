-- name: ListProducts :many
SELECT * FROM products ORDER BY id;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1;

-- name: ListAllProducts :many
SELECT * FROM products ORDER BY id;

-- name: GetProductsByIDs :many
SELECT * FROM products WHERE id = ANY($1::text[]);
