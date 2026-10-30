-- +goose Up
-- Payee rules: "this payee is a <category>". New spending whose merchant contains payee_norm is filed under
-- the rule's category when no category is given. payee is what the user said; payee_norm is what is matched.
CREATE TABLE payee_rules (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payee       text NOT NULL,
    payee_norm  text NOT NULL,
    category    text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, payee_norm)
);

-- +goose Down
DROP TABLE payee_rules;
