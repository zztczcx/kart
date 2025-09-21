-- name: GetCoupon :one
SELECT * FROM coupons WHERE code = $1;
