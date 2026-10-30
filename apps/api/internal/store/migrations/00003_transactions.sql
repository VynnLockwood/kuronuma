-- +goose Up
-- Plutus: spending records. Amounts are stored as hundredths of the currency unit.
CREATE TABLE transactions (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    currency     text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    merchant     text NOT NULL,
    category     text NOT NULL DEFAULT 'uncategorized',
    occurred_on  date NOT NULL,
    note         text NOT NULL DEFAULT '',
    reference    text,
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- The same slip (same reference) cannot be saved twice for one user.
CREATE UNIQUE INDEX transactions_user_reference_idx
    ON transactions (user_id, reference) WHERE reference IS NOT NULL;
CREATE INDEX transactions_user_date_idx ON transactions (user_id, occurred_on DESC);

-- +goose Down
DROP TABLE transactions;
