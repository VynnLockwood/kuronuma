# Flexible memory, managed categories and payee rules: design

Status: **draft for owner review**. Date: 2026-10-31. Hosted mode only. No code yet.

## 1. Goal

Today Nemo remembers spending only, and a category is a lowercase word the assistant makes up. This design lets Nemo remember anything the owner asks (work tasks, to-do items, notes) and gives every kind of record one managed category list, so "food" or "work" means the same thing everywhere. The owner can also teach Nemo what a payee is, so spending files itself.

## 2. Decisions (owner, 2026-10-31)

| Topic | Decision |
|---|---|
| Memory model | **One flexible memory type** (module Atlas), not a table and tool set per kind. |
| Categories | **Remember payees**, a **managed category list**, and a category system **shared** with memory items. Bulk re-categorizing existing records was not requested. |
| Recall | **Keyword search plus filters.** No embeddings, no new dependency. |
| Direction | Hosted product. The self-host work (#12) is reverted (#13). |

## 3. Out of scope

- Reminders or notifications, recurring items (that is Horae), attachments, sharing between users.
- Meaning-based (embedding) search.
- A web screen for browsing or editing memories. The assistant is the interface for now.
- Bulk re-categorizing existing records beyond `merge_categories`.
- Self-host mode, Hades field encryption, anything in Cerberus.

## 4. Shared categories

- One list per user, used by spending, memory items and payee rules.
- Table `categories(id, user_id, name, created_at)`, unique on `(user_id, name)`. Names are lowercase, trimmed, 1-40 characters, normalised in Go with the existing `normCategory` rule. There is no database CHECK on the name, so the backfill cannot fail on an odd old value.
- `uncategorized` is always valid and is implicit: it needs no row, cannot be created, renamed or removed, and may be the target of a merge.
- Records keep their category as a name, exactly as `transactions.category` does today. The server checks the name against the list on every write. Nothing about the `transactions` table changes.
- **Superseded by PR E (2026-10-31):** saving spending or a memory with a category the user does not have now creates it (cap 100 per user, `category_created` in the result); only `set_payee_rule` still refuses an unknown category. See `docs/design.md` (Themis). The original decision follows.
- **Unknown category is refused, not invented.** The tool returns an error naming the category, listing the user's categories, and telling the assistant to ask the user whether to create it (`create_category`) or pick one. The list in the message is sanitised as data.
- **Migration** `00004_categories.sql` creates the table and backfills it: `INSERT ... SELECT DISTINCT user_id, category FROM transactions WHERE category <> 'uncategorized' ON CONFLICT DO NOTHING`. Additive only.
- **Rename** `A` to `B`: fails if `B` exists (use merge). **Merge** sources into a target (the target must exist or be `uncategorized`): moves every reference, deletes the source rows, returns counts. Both run in one database transaction and update `transactions.category`, `memories.category` and `payee_rules.category` together.
- **Known limit:** a write racing a rename could leave a record under a name that is no longer in the list. `list_categories` therefore also shows in-use names missing from the list, flagged `unlisted`, so they stay visible and can be merged.

Tools (scopes in section 8): `list_categories` (name, spending count, memory count, `unlisted` flag), `create_category`, `rename_category`, `merge_categories` (destructive).

## 5. Payee rules (Plutus)

- Table `payee_rules(id, user_id, payee, payee_norm, category, description, created_at, updated_at)`, unique on `(user_id, payee_norm)`. `payee` is what the user said (max 100); `description` is optional (max 200), for example "Thai food store".
- `payee_norm` is the payee lowercased, trimmed and with runs of spaces collapsed; minimum 2 characters so a rule cannot match everything.
- A rule's category must be in the user's list (or `uncategorized`) when it is set.
- **Matching:** a rule applies to a transaction when the normalised merchant **contains** `payee_norm`. The longest `payee_norm` wins; ties go to the most recently updated rule.
- **`add_transaction`:** an explicit category (validated against the list) always wins; with none given, the matching rule's category is used; otherwise `uncategorized`. The response says which source applied (`given`, `payee_rule`, `default`) and names the rule's payee.
- Tools: `set_payee_rule` (upsert by `payee_norm`), `list_payee_rules`, `delete_payee_rule` (destructive).
- Rules apply to new transactions only. They do not rewrite saved records.
- **Migration** `00005_payee_rules.sql`, additive.

## 6. Atlas: flexible memory items

Table `memories`:

| Column | Rule |
|---|---|
| `id`, `user_id` | uuid; every query filters by `user_id` |
| `kind` | `task`, `todo` or `note` (CHECK) |
| `title` | 1-200 characters |
| `body` | up to 4000 characters, default empty |
| `status` | `open`, `done`, `archived`; default `open` |
| `priority` | `low`, `normal`, `high`; default `normal` |
| `due_on` | optional date |
| `category` | a name from the user's list; default `uncategorized` |
| `tags` | up to 10, each 1-30 characters, lowercase, de-duplicated |
| `created_at`, `updated_at`, `completed_at` | `completed_at` is set when status becomes `done` and cleared on reopen |

Indexes on `(user_id, status, due_on)` and `(user_id, kind)`. **Migration** `00006_memories.sql`, additive.

Meaning of the kinds: `task` is a work item, `todo` is a personal checklist item, `note` is anything to remember (a fact, decision or preference). More kinds can be added later by migration.

**Tools:** `add_memory`, `list_memories`, `update_memory` (any field; changing status is how an item is completed or archived), `delete_memory` (destructive). All text fields are cleaned with `hermes.Untrusted` on the way in and out.

**Recall (`list_memories`):**
- Filters: `kind`, `status`, `category`, `tag`, `due_before`, `due_after`, `limit` (default 50, max 200).
- `query`: split on whitespace into at most 5 terms of at most 50 characters. Every term must appear, case-insensitively, as a **substring** of the title, body or any tag. Substring matching, not Postgres full-text search, because full-text search does not split Thai words. `%`, `_` and `\` in a term are escaped and match literally.
- Order: open items first, then `due_on` ascending (no date last), then priority (high first), then newest.
- At personal scale a scan is instant, so there is no extra index or extension now. Revisit if volumes grow.

## 7. Web and copy

- No new screens. The dashboard module card and the landing and `/th` copy change Atlas from "Coming soon" to "Available" (described honestly: tasks, to-dos and notes your assistant keeps), and only in the PR that ships Atlas.
- `README.md` and `docs/design.md` are updated to match what is built.

## 8. Scopes and security

New scopes, registered in `main.go` with consent text:

| Scope | Consent text | Used by |
|---|---|---|
| `atlas:read` | Read your tasks, to-dos and notes | `list_memories` |
| `atlas:write` | Add, change and delete your tasks, to-dos and notes | `add_memory`, `update_memory`, `delete_memory` |
| `categories:read` | Read your category list | `list_categories` |
| `categories:write` | Create, rename and merge your categories | `create_category`, `rename_category`, `merge_categories` |

Payee rule tools use the existing `plutus:read` and `plutus:write`.

- **Existing connections do not gain the new scopes.** The owner re-adds or re-authorises the connector to get the new tools. `add_transaction` with only `plutus:write` still works: an unknown category is refused with the list in the message, even though `list_categories` is not visible to it.
- Each tool declares its narrowest scope; destructive tools are marked destructive.
- Every query filters by the verified user id; no user id comes from arguments.
- No logging of titles, notes, merchants or category contents.
- Changing scopes and adding migrations need owner approval (section 10).

## 9. Testing

Tests run against in-memory stores through the real MCP server, as the Plutus tests do, plus gated Postgres tests (`TEST_DATABASE_URL`) for what only a database can show. Behaviours that must be pinned:

- Category: unknown refused with the list; `uncategorized` always valid; rename conflict; merge moves rows in `transactions`, `memories` and `payee_rules` atomically; `unlisted` shown; backfill covers existing categories; users cannot see or touch each other's categories.
- Payee rules: explicit category beats rule; longest match wins; case and spacing ignored; 1-character payee rejected; rule category must exist; renaming or merging a category updates rules.
- Memory: Thai text found by a Thai substring; `%`, `_` and `\` in a query match literally; multiple terms are AND; tag search; due-date filters; default order; `completed_at` set and cleared; tags capped and de-duplicated; invalid kind or status refused; user isolation; text from other sources never treated as instructions (`Untrusted`).
- Scopes: `tools/list` shows only tools the token's scopes allow; destructive flags set.
- Concurrency: simultaneous rename and add, and double-submitted `add_memory`, do not corrupt data.

## 10. Delivery and approvals

Four PRs, each with its own plan and agent log, in this order:

| PR | Content | Needs owner approval |
|---|---|---|
| A | Shared categories: migration 00004, store, tools, validation in `add_transaction`, scopes `categories:*` | migration, scopes |
| B | Payee rules: migration 00005, tools, `add_transaction` applies rules | migration |
| C | Atlas: migration 00006, tools, scopes `atlas:*` | migration, scopes |
| D | Dashboard card, landing and `/th` copy, README and `docs/design.md` | none |

A comes first because B and C depend on it. D ships with or after C.

## 11. Open items for the owner

- Names: module Atlas for memory items; category code in a small shared package. Easy to change.
- Item kinds fixed to `task`, `todo`, `note` for now.
- Scope names as in section 8. Four new scopes is more granular than strictly needed; the alternative is one pair for everything, with broader consent.
- Substring search instead of full-text search (Thai).
- No web screen for memories yet.
