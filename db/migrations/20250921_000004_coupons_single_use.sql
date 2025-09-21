-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS coupon_redemptions (
  code TEXT PRIMARY KEY REFERENCES coupons(code) ON DELETE CASCADE,
  redeemed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS coupon_redemptions;
-- +goose StatementEnd


