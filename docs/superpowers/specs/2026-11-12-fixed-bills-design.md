# Fixed monthly bills (Plutus) design

Status: approved in brainstorming on 2026-11-12, pending review of this written spec.

## Purpose

Let the user record fixed monthly costs (subscriptions, rent, mobile bills), see one monthly total for them, and confirm each month's payment. Confirming a payment creates a normal spending record, so spending history stays complete. Everything works through MCP tools; the dashboard calls the same operations.

## Decisions (from brainstorming)

- Bills live inside the Plutus module. They use the existing `plutus:read` and `plutus:write` scopes and the shared category list.
- A confirmed payment creates a spending record linked to its bill.
- Changing a bill's amount changes only the bill. Past payments keep the amount actually paid.
- Ending a bill sets the month from which it stops counting toward totals. Past months still show it.
- `delete_bill` is refused when the bill has confirmed payments; the user ends the bill instead.
- No reminders or due-day alerts in this version. Those belong to Horae later.

## Data model

Migration `00007_bills.sql` adds one table, `bills`:

- `id`: same id style as the existing tables.
- `user_id`: owner; every query filters on it.
- `name`: free text, sanitised like other Plutus text fields.
- `amount`: `bigint` in hundredths of the currency unit.
- `currency`: one of the currencies Plutus already accepts.
- `category`: a name from the shared list (Themis); `uncategorized` is valid.
- `start_month`: first month (stored as the first day of the month) in which the bill counts.
- `ended_from`: optional first month in which the bill no longer counts.
- `created_at`, `updated_at`.

Payments need no table. A payment is a transaction whose `reference` is `bill:<bill_id>:<YYYY-MM>`. The existing unique index on `(user_id, reference)` blocks a second confirmation for the same month.

## Rules

- **Counting:** a bill counts in month M when `start_month <= M` and (`ended_from` is empty or `M < ended_from`).
- **Monthly total:** the sum of planned amounts of the bills that count in M, one total per currency. Currencies are never mixed.
- **Paid status:** a bill is paid for month M when a transaction with reference `bill:<bill_id>:<YYYY-MM>` exists. Deleting that transaction makes the month unpaid again.
- **Confirming a payment** creates a transaction with:
  - merchant: the bill's name;
  - category: the bill's category;
  - amount: the planned amount unless the caller gives the amount actually paid;
  - date: the date given, or today if none is given. The date must fall within the month being confirmed, otherwise the call is refused (see "Choices made in this spec").
  - reference: `bill:<bill_id>:<YYYY-MM>`.
- **Refused confirmation:** a second confirmation for the same bill and month is refused with a message that says the month is already paid.
- **Amount changes** update only `bills`. Transactions are never changed.
- **Ending** sets `ended_from`. Clearing `ended_from` through `update_bill` undoes the ending.
- **Deleting** is refused when any confirmed payment exists for the bill.
- **Amounts** follow the existing Plutus rules: at most two decimals, positive, rejected rather than rounded when too precise.
- **Cap:** 100 bills per user. A new bill past the cap is refused with a message asking the user to end or delete one first.

## MCP tools

All tools use the Plutus scopes. Existing Plutus connectors see the new tools without re-authorising. Every tool rejects unknown arguments.

| Tool | Scope | Purpose |
|---|---|---|
| `list_bills` | `plutus:read` | Bills that count in a month (default: current month). Each has planned amount, currency, category, paid flag, and the amount and date paid. |
| `monthly_bills_summary` | `plutus:read` | For one month, per currency: planned total, paid total, number of paid and unpaid bills. |
| `add_bill` | `plutus:write` | Name, amount, currency, category, start month (default: current month). Creates a new category if needed. |
| `update_bill` | `plutus:write` | Change name, amount, category, start month, or clear `ended_from`. |
| `end_bill` | `plutus:write` | Set `ended_from` to a month. |
| `confirm_bill_payment` | `plutus:write` | Create the linked spending record for one bill and month. Optional amount and date. |
| `delete_bill` | `plutus:write`, destructive | Delete a bill with no confirmed payments. |

Tool descriptions tell the assistant to show the planned amount and ask for confirmation before confirming a payment, as the slip flow does for spending.

## Dashboard

- A **Bills** tab on the spending page, with a month selector.
- At the top: the planned and paid totals per currency (`monthly_bills_summary`).
- Below: each bill that counts in the month, with its planned amount and a Paid or Unpaid status. A **Mark paid** button calls `confirm_bill_payment` with the planned amount.
- Add, edit and end actions. Delete appears only for bills with no confirmed payments.
- Calls go to the Go API under `/api/web/`, through the same operations as the MCP tools. The web app holds no business logic.
- Thai and English strings; existing design tokens; no decorative motifs on this data page (an empty list may show one, as the design rules allow).

## Errors

Errors follow the existing pattern: a 422 with a message written for people when a tool refuses input, and a generic 500 otherwise. Messages name the bill and month, use the shared gentle voice, and never include user text unescaped.

## Testing

- **Totals:** per currency, with bills excluded from the month they end and from months before they start. Test the month edges.
- **Payments:** default amount, custom amount, date defaulting to today, date outside the month refused, second confirmation refused.
- **History:** changing a bill's amount leaves past transactions unchanged.
- **Deletion:** refused when payments exist; allowed otherwise.
- **Access:** a `plutus:read` token cannot call write tools and does not see them in `tools/list`. One user cannot see or confirm another user's bills.
- **Limits:** the 100-bill cap; unknown arguments refused.
- **Database:** migration 00007 runs up, down and up again against a real Postgres.
- **Web:** type-check, `npm run check:design`, and a browser smoke test of the Bills tab.

## Choices made in this spec

These are not in the approved brainstorming answers. Review them:

1. A confirmed payment's date must fall within the month being confirmed.
2. A bill with no category uses `uncategorized`.
3. `ended_from` and `start_month` are stored as the first day of the month, and a month is given as `YYYY-MM` in tools.

## Out of scope

- Reminders, due days, and recurring transaction generation (Horae, later).
- Price history per bill (rejected in favour of keeping only the current amount).
- Bank connections or automatic payment detection.

## Self-review

- Placeholders: none.
- Consistency: the counting rule, the paid-status rule and the delete rule agree with each other and with the approved decisions.
- Scope: one module, one table, seven tools, one dashboard tab. Small enough for one implementation plan.
- Ambiguity: the three choices above are the only places the spec adds rules beyond the approved answers.
