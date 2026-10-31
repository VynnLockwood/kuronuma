-- +goose Up
-- Atlas: things the user's assistant remembers for them (tasks, to-dos, notes). The category is a name from the
-- user's shared category list (checked by the server, as for transactions). Search is substring matching in the
-- application's queries, not full-text search, so it works for Thai.
CREATE TABLE memories (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind         text NOT NULL CHECK (kind IN ('task', 'todo', 'note')),
    title        text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    body         text NOT NULL DEFAULT '' CHECK (char_length(body) <= 4000),
    status       text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'done', 'archived')),
    priority     text NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high')),
    due_on       date,
    category     text NOT NULL DEFAULT 'uncategorized',
    tags         text[] NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

CREATE INDEX memories_user_status_due_idx ON memories (user_id, status, due_on);
CREATE INDEX memories_user_kind_idx ON memories (user_id, kind);

-- +goose Down
DROP TABLE memories;
