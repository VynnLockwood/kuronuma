-- +goose Up
-- Shared category list. Records (transactions today, more later) refer to a category by name; the server
-- checks the name against this list when saving. "uncategorized" is built in and has no row.
-- There is deliberately no CHECK on name: the rules live in Go, so the backfill below cannot fail on an old value.
CREATE TABLE categories (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

-- Everyone's existing categories become their starting list, so nothing they already use is refused.
INSERT INTO categories (user_id, name)
SELECT DISTINCT user_id, category FROM transactions WHERE category <> 'uncategorized'
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE categories;
