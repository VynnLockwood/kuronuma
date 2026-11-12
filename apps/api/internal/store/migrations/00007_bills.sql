-- +goose Up
-- Fixed monthly bills. A bill counts in every month from start_month up to, but not including, ended_from.
-- Months are stored as first-day-of-month dates. Payments are ordinary transactions whose reference is
-- bill:<id>:<YYYY-MM>.
CREATE TABLE bills (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         text NOT NULL,
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency     text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    category     text NOT NULL,
    start_month  date NOT NULL,
    ended_from   date,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CHECK (ended_from IS NULL OR ended_from >= start_month)
);
CREATE INDEX bills_user_idx ON bills (user_id);

-- +goose Down
DROP TABLE bills;
