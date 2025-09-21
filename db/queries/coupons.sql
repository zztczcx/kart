-- name: GetCoupon :one
SELECT * FROM coupons WHERE code = $1;

-- name: TryRedeemSingleUse :one
INSERT INTO coupon_redemptions (code)
VALUES ($1)
ON CONFLICT DO NOTHING
RETURNING code;
