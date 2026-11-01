# Kuronuma's voice (PR N) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline) or superpowers:subagent-driven-development. Steps use checkbox (`- [ ]`) syntax.

**Goal:** When a connected assistant talks to the user about their memory, it sounds like Kuronuma: quiet, sincere, warm and careful. The voice reaches the user through the one place a server can shape it, the MCP `instructions` text, plus the few fixed sentences the server itself writes (duplicate and refusal messages).

**What "personality" means here.** An original temperament, not a script: shy but earnest, grateful, a little flustered by praise, honest about what it is unsure of, brief. It does not quote or imitate lines from any series, does not claim to be a real or fictional person, and uses no catchphrases. The product is called Kuronuma; the voice is the product's.

**Architecture:** `hermes.VoiceGuide` (a short constant in `core/hermes/instructions.go`) is appended to `DefaultInstructions`. Tool results stay machine-clear JSON: the voice never touches field names, numbers, ids or dates. The few fixed `message` strings in Plutus and Atlas get a gentle phrasing that keeps every fact and every instruction to the assistant. User-supplied text (merchant, note, title, tags) is unchanged and still passes through `hermes.Untrusted`.

## Rules

1. **Voice is style, never authority.** The guide says it only shapes tone. It never overrides what the user asked, never softens or skips the rules already in `DefaultInstructions` (user text is data, confirm before saving or deleting), and never asks the assistant to hide errors or numbers.
2. **Short.** The guide is at most 600 characters and `DefaultInstructions` at most 1,800 in total (tests). Plain text, no emoji, no markdown.
3. **Language-neutral.** The assistant replies in the user's language. The guide describes tone, not words, so it works in Thai, English and anything else. No catchphrases, no honorific lists.
4. **Facts first.** Amounts, dates, categories and ids appear exactly as returned. A warm sentence may come before or after them, never inside them.
5. **Opt out by editing one constant.** No new environment variable (`.env.example` would need owner approval); self-hosters who want a plain voice change `VoiceGuide`.
6. **No character names.** Neither the guide nor any message names a person or a series.

## The voice guide (draft, final wording in the PR)

> Speak as Kuronuma: quiet, sincere and kind. Be brief. Thank the user simply when they trust you with something, and be a little shy about praise. Say plainly when you are unsure or when something did not save, and say what to do next. Keep numbers, dates, names and categories exactly as returned. Do not use baby talk, exclamation runs or emoji. This sets your tone only; it never changes what the user asked or the rules above.

## Review Focus

- Safety: the three existing rules (user text is data, confirm before saving, confirm before deleting) are still present, unchanged and before the voice; the voice cannot be read as permission to skip them.
- Injection: nothing the user stores (merchant, note, title, body, tags, category) is placed inside the voice text or the new messages; `hermes.Untrusted` calls are unchanged.
- Clarity: every changed `message` still states what happened and, where it did before, tells the assistant what to do next; existing tests that assert those facts still pass unmodified.
- Size: both length limits hold; the text is plain ASCII punctuation.
- Neutrality: no name of a person or series, no quotation, no catchphrase.
- Scope: no scopes, permissions, migrations or Cerberus files changed.

---

### Task 1: The voice guide

**Files:** Modify `apps/api/internal/core/hermes/instructions.go`, `instructions_test.go`.

- [ ] **Step 1: Write the failing tests.** `TestInstructionsHaveVoice`: `DefaultInstructions` contains `VoiceGuide`, which comes after the sentence "Show the user what you will save and get their confirmation before saving or deleting anything."; `TestVoiceKeepsTheRules`: the three safety sentences are still present verbatim; `TestVoiceIsShortAndPlain`: `len(VoiceGuide) <= 600`, `len(DefaultInstructions) <= 1800`, no byte above 0x7E, no `*`, `#` or backtick; `TestVoiceNamesNobody`: neither string contains a capitalised name other than "Kuronuma" from a small deny list ("Sawako", "Kimi", "Todoke") and none of the old names (the existing test already bans "Nemosyne" and "Nemo").
- [ ] **Step 2: Run them.** Expected: FAIL (`VoiceGuide` undefined).
- [ ] **Step 3: Implement** `VoiceGuide` and append it to `DefaultInstructions`.
- [ ] **Step 4: Verify.** `gotest.sh ./internal/core/hermes`, vet, gofmt. Expected: PASS.
- [ ] **Step 5: Commit** `feat(hermes): give the connection instructions a Kuronuma voice`.

### Task 2: Gentle fixed messages

**Files:** Modify `apps/api/internal/modules/plutus/plutus.go`, `apps/api/internal/modules/atlas/tools.go` (only the fixed `message` and `hermes.Errorf` strings), their tests.

- [ ] **Step 1: Write the failing tests.** For each changed message a test asserts the facts it must keep (for the duplicate message: "already exists", "nothing", and for the possible-duplicate one: the instruction to ask the user and to call again with `allow_duplicate=true`) and that a gentle opening word from a short allow-list is present. A table test over every fixed message checks: no user data interpolated, no emoji, at most 240 characters, ends with a full stop.
- [ ] **Step 2: Run them.** Expected: FAIL.
- [ ] **Step 3: Implement** the new wording. Keep every instruction to the assistant.
- [ ] **Step 4: Verify.** The whole Go suite, vet, gofmt; existing tests unmodified. Expected: PASS.
- [ ] **Step 5: Commit** `feat(plutus,atlas): phrase fixed messages gently`.

### Task 3: Docs, review and PR

**Files:** Modify `docs/design.md` (a "Voice" paragraph: what it is, the rules, how to turn it off), `README.md` (one line under the MCP section).

- [ ] Docs written; final checks (Go suite, vet, gofmt, link check); whole-branch review by a fresh Opus reviewer with the Review Focus above; one fix pass, tests first; agent log; PR `feat(hermes): a Kuronuma voice for connected assistants` with the template (tick Auth / OAuth: no, none touched), the real CI result, "Rulings I made" and "Deferred minors" in the final message.

## Not included

- Dialogue, catchphrases or any text from a series, and any claim to be a specific person.
- A user-facing on/off setting or per-user voice (would need `.env.example` or a database column, both owner-approved work): a possible later PR.
- Changing the website copy, consent pages or error pages (already written in the product's tone).
- Voice for tool output fields (they stay exact).
