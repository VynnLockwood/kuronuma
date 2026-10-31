# Kuronuma rename (PR R) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rename Nemosyne and Nemo to Kuronuma across the code, copy, identifiers, build files and docs, with the look unchanged.

**Architecture:** A mechanical rename guarded by a brand check that fails until nothing old is left. Four independent tasks, each with its own tests: Go module and build, assistant-facing text and identifiers, the web app, deploy files and docs. The database name and user and Compose's project name stay `nemosyne` on purpose.

**Tech Stack:** Go 1.26 API, Next.js 16 with Tailwind and shadcn-style components, nginx, Docker Compose. No new dependencies.

**Spec:** `docs/superpowers/specs/2026-10-31-kuronuma-rebrand-design.md` (section 1). The domain, Cloudflare, Google sign-in and `PUBLIC_URL` steps (section 4) are done by the owner and another agent, not in this PR.

## Global Constraints

- Names: public name **Kuronuma**, Thai **คุโรนุมะ**, repository `kuronuma-mcp`, MCP server `Kuronuma`, Go module `github.com/VynnLockwood/kuronuma-mcp/apps/api`.
- Stays `nemosyne`: database name, database user, Docker Compose `name:` (the data volume is named after it; renaming it orphans the data). Add a comment in `deploy/docker-compose.yml` saying why.
- Do not edit `agent-logs/**` (except the new log and `agent-logs/README.md`) or `docs/superpowers/**` (history).
- The logo image stays as it is (file renamed only). No new dependencies. No look changes (V1 does that).
- Cookies `nem_session` and `nem_flow` become `kuro_session` and `kuro_flow`; client id prefix `nem_c_` becomes `kuro_c_` for new clients, and stored `nem_c_` ids keep working; HMAC domain strings become `kuronuma-...`.
- Commits: `VynnLockwood <120536136+VynnLockwood@users.noreply.github.com>`, no AI trailers. Conventional Commits. Branch `refactor/kuronuma-rename`.
- Owner approval already needed and given at plan approval: `deploy/`, Dockerfiles, CI, sign-in code (cookie names, consent text). Say so in the PR.

## Review Focus

- The Compose project and volume name must not change: `docker compose config` must still show project `nemosyne` and volume `nemosyne_gaia`.
- A request carrying only the old `nem_session` cookie is signed out (no fallback), while `kuro_session` works.
- A client stored with an old `nem_c_` id still reaches the consent page; a newly registered client gets `kuro_c_`.
- No page, title, alt text, meta tag or connect command in the web app still says Nemo, Nemosyne or เนโม่, in English or Thai.
- The API image and nginx container still build and start with the renamed `cmd/kuronuma-api` and `kuronuma-proxy.conf` paths (CI builds both).

---

### Task 1: Brand check (the failing test for the whole rename)

**Files:**
- Create (outside the repo): `/home/claude/scratch/check-brand.sh`

**Interfaces:**
- Produces: `check-brand.sh` exits 1 and prints every offending line while old names remain; exits 0 when only allowed lines remain.

- [ ] **Step 1: Write the script.** Run from the repo root:
  `git grep -I -i -n -E 'nemosyne|\bnemo\b|nemo-|เนโม่|nem_(session|flow|c_)' -- . ':!agent-logs' ':!docs/superpowers' ':!apps/web/package-lock.json'` and drop lines that are allowed: every line of `deploy/docker-compose.yml`, and any line containing the marker `[keep-nemosyne]` (put in a comment or sentence that deliberately keeps an old name). Exit 1 if anything is left, and print the count. Also fail if `apps/web/package-lock.json` has a `"name": "nemosyne-web"`.
- [ ] **Step 2: Run it.** Expected: FAIL, around 230 lines in about 55 files.

### Task 2: Go module path, command and build files

**Files:**
- Modify: `apps/api/go.mod`; every `apps/api/**/*.go` import of `github.com/VynnLockwood/nemosyne/apps/api/...`; `apps/api/Dockerfile` (build path `./cmd/kuronuma-api`, binary name); `.github/workflows/ci.yml` (image names `kuronuma-api:ci`, `kuronuma-web:ci`); `/home/claude/scratch/gotest.sh` (module path and the `cmd` dir if it copies it).
- Move: `git mv apps/api/cmd/nemosyne-api apps/api/cmd/kuronuma-api`.

**Interfaces:**
- Produces: module path `github.com/VynnLockwood/kuronuma-mcp/apps/api`, used by every later task's imports.

- [ ] **Step 1: Rename.** `go.mod` module line, a scripted replace of the import prefix (only the prefix, nothing else), the directory move, and the Dockerfile and CI strings. Leave comments for Task 3.
- [ ] **Step 2: Format.** `gofmt -w` the touched files (import order can change); `gofmt -l apps/api` must print nothing.
- [ ] **Step 3: Verify.** `/home/claude/scratch/gotest.sh vet` and `/home/claude/scratch/gotest.sh` pass (cerberus, hermes, themis, plutus, atlas). `git grep -n "nemosyne/apps/api" -- apps/api` returns nothing. `internal/store` and `cmd` compile only in CI.
- [ ] **Step 4: Commit** `refactor(api): rename the Go module and command to kuronuma`.

### Task 3: Assistant-facing text and identifiers

**Files:**
- Modify: `apps/api/internal/core/hermes/instructions.go`, `apps/api/internal/core/hermes/server.go` (comment), `apps/api/cmd/kuronuma-api/main.go` (`Name: "kuronuma"`, `Title: "Kuronuma"`), `apps/api/internal/core/cerberus/{handler.go,flow.go,oauth.go,consent.go,types.go}`, tool descriptions and comments in `modules/plutus`, `modules/atlas`, `core/themis`, `internal/store`, `config`; tests in the same packages.
- Test: `cerberus/handler_test.go`, `cerberus/oauth_test.go`, `cerberus/consent_test.go`, `cerberus/oauth_gemini_test.go`, `hermes/instructions_test.go`.

**Interfaces:**
- Consumes: Task 2's module path.
- Produces: `sessionCookie = "kuro_session"`, `flowCookie = "kuro_flow"`, new client ids `"kuro_c_" + id`, `flowDomain = "kuronuma-oauth-flow-v1:"`, `consentDomain = "kuronuma-oauth-consent-v1\x00"`.

- [ ] **Step 1: Write the failing tests** (names and assertions):
  - `TestOldSessionCookieNameIsNotAccepted`: sign in through the harness, then `GET /api/auth/me` with the same cookie value named `nem_session` returns 401, and named `kuro_session` returns 200.
  - `TestNewClientsGetTheKuroPrefix`: a registered client's id has prefix `kuro_c_`.
  - `TestLegacyClientIDsStillWork`: a client stored with id `nem_c_legacy` (line carries `[keep-nemosyne]`) reaches the consent page with status 200 and can complete the authorization-code exchange.
  - `TestDefaultInstructions` expects "Kuronuma is the user's memory" and no "Nemosyne" or "Nemo".
  - Consent tests: English heading `Connect Google to your Kuronuma?`, check text `Approve if you just added Kuronuma as a connector in Gemini, Claude or ChatGPT.`; Thai heading `เชื่อมต่อ Google กับคุโรนุมะของคุณไหม?`; page title `Connect to Kuronuma` and `เชื่อมต่อกับคุโรนุมะ`; header wordmark `KURONUMA`; logo `<img src="/brand/kuronuma-logo.png"`; `TestErrorPageIsBranded` expects `KURONUMA`.
- [ ] **Step 2: Run them.** Expected: FAIL for those reasons.
- [ ] **Step 3: Implement the renames** listed under Interfaces, the consent and error copy (English: "Connect %s to your Kuronuma?", "Approve if you just added Kuronuma as a connector...", `PageTitle` "Connect to Kuronuma"; Thai uses คุโรนุมะ in running text), and every remaining "Nemosyne" or "Nemo" in Go comments and tool descriptions (for example "Teach Nemo what a payee is" becomes "Teach Kuronuma what a payee is"). Update existing tests that assert the old text or the `nem_c_nope` ids (use `kuro_c_nope`).
- [ ] **Step 4: Verify.** gotest.sh and vet pass; `gofmt -l apps/api` empty; `check-brand.sh` shows no remaining lines under `apps/api`.
- [ ] **Step 5: Commit** `refactor(api): rename the public name, cookies and client id prefix to kuronuma`.

### Task 4: Web app

**Files:**
- Modify: `apps/web/lib/brand.ts` (`name: "Kuronuma"`, `thai: "คุโรนุมะ"`, `full: "Kuronuma"`; drop the "friendly name" comment), `apps/web/components/landing/copy.ts` (English and Thai), `apps/web/app/layout.tsx` (title template `%s | Kuronuma`, `applicationName`, `siteName`, og alt), `apps/web/app/connect/{layout.tsx,page.tsx}`, `apps/web/app/dashboard/page.tsx`, `apps/web/components/{Logo.tsx,landing/Landing.tsx}`, `apps/web/app/globals.css` (comment), `apps/web/package.json` and the two name fields in `package-lock.json` (`kuronuma-web`).
- Move: `git mv apps/web/public/brand/nemo-logo.png apps/web/public/brand/kuronuma-logo.png`; delete `nemo-mark.svg` and `nemo-mark.png` (unused); regenerate `apps/web/public/og.png` (1200 by 630) with the logo, "Kuronuma" and the tagline, using a Playwright screenshot script kept outside the repo.
- Test (outside the repo): `/home/claude/scratch/brand-check.py`.

**Interfaces:**
- Produces: `BRAND.name`, `BRAND.thai`, logo path `/brand/kuronuma-logo.png`.

- [ ] **Step 1: Write the failing browser test** `brand-check.py`: for `/`, `/th`, `/connect`, `/dashboard` (API calls faked: `/api/auth/me` 401, and once 200) at 1280 and 390 px wide, the visible text, `<title>`, meta tags, every `alt` and every image `src` contain none of `Nemo`, `Nemosyne`, `เนโม่`; `/connect` shows `claude mcp add --transport http kuronuma`; the header wordmark reads `Kuronuma` and the page has no horizontal overflow; `/brand/kuronuma-logo.png` and `/og.png` (1200 by 630) load. Expected now: FAIL.
- [ ] **Step 2: Rewrite the copy.** Decisions the code cannot show:
  - Replace "Nemo" and "Nemosyne" with "Kuronuma" in English, and "เนโม่" with "คุโรนุมะ" in Thai (so "เชื่อมต่อเนโม่แล้ว" becomes "เชื่อมต่อคุโรนุมะแล้ว").
  - Titles: English `Kuronuma (คุโรนุมะ): a memory for your AI assistant`; Thai `คุโรนุมะ (Kuronuma): ความจำสำหรับผู้ช่วย AI ของคุณ`.
  - Remove the "short for Nemosyne" clauses (hero sub, meta description, both FAQ answers, Thai equivalents); hero sub starts `Kuronuma gives any assistant that supports MCP a memory...` and Thai `คุโรนุมะให้ผู้ช่วยที่รองรับ MCP มีความจำ...`.
  - FAQ English: question `What is Kuronuma?`, answer starts `Kuronuma is an open-source memory for AI assistants.` and keeps the rest; question `Why “Kuronuma”?`, answer `It's a name we like: quiet, gentle and good at listening, which is what a memory should be. In Japanese it also means “black marsh”, a still place where things settle.` Thai question `ทำไมถึงชื่อคุโรนุมะ`, answer `เป็นชื่อที่เราชอบ เงียบ อ่อนโยน และเป็นผู้ฟังที่ดี เหมือนที่ความจำควรเป็น ในภาษาญี่ปุ่นแปลว่า “หนองน้ำสีดำ” ที่นิ่งสงบและทุกสิ่งตกตะกอนอยู่`.
  - Footer: `blurb` "A memory for your AI assistant. Open source." / Thai "ความจำสำหรับผู้ช่วย AI ของคุณ โอเพนซอร์ส"; `built` "Kuronuma is open source (MIT)." / "คุโรนุมะเป็นโอเพนซอร์ส (MIT)".
  - Modules section: title `Modules` / `โมดูล`; sub `Each kind of memory is a small module. Spending and tasks are live; more are on the way.` / `ความจำแต่ละประเภทเป็นโมดูลเล็ก ๆ ค่าใช้จ่ายและงานใช้งานได้แล้ว และจะมีอีกหลายอย่างตามมา`. Module names (Plutus, Atlas, Horae, Hephaestus) stay as labels.
  - Connect page: command `claude mcp add --transport http kuronuma <url>`, "Your Kuronuma MCP URL".
- [ ] **Step 3: Rename the files and update the logo paths and package names** as listed; regenerate `og.png`.
- [ ] **Step 4: Verify.** `npx tsc --noEmit`, `npm run check:design`, `npm run build` pass; `brand-check.py` passes; screenshots of `/` and `/th` at 390 px show the longer wordmark fits the header.
- [ ] **Step 5: Commit** `refactor(web): rename the site to Kuronuma`.

### Task 5: Deploy files and docs

**Files:**
- Modify: the file `deploy/nginx/proxy.conf` keeps its name, but the path it is mounted at inside the container changes from `nemosyne-proxy.conf` to `kuronuma-proxy.conf`: in `deploy/nginx/nginx.conf` (seven includes), the mount in `deploy/docker-compose.yml`, and the CI `nginx -t` mount. Also: `deploy/docker-compose.yml` (comment above `name: nemosyne`), `README.md`, `AGENTS.md`, `SECURITY.md`, `agent-logs/README.md`, `docs/design.md`, `.env.example` only if it names the project.

**Interfaces:**
- Consumes: Tasks 2 to 4.

- [ ] **Step 1: Rename the include and add the comment.** The compose comment reads: `The project name, database name and user stay "nemosyne" on purpose: the data volume is named after the project, so renaming it would leave existing data behind.` Mark any other deliberately kept old name with `[keep-nemosyne]`.
- [ ] **Step 2: Rewrite the docs.** README title `# Kuronuma` and first paragraph `**Kuronuma (คุโรนุมะ)** is an open-source memory for AI assistants...` (drop the Mnemosyne sentence); `claude mcp add --transport http kuronuma https://your-domain/mcp`; `go run ./cmd/kuronuma-api`. `docs/design.md` line "Names" becomes: `Kuronuma (คุโรนุมะ) is the name used in the UI, copy, repository (kuronuma-mcp), MCP server and cookies. The database name, database user and Compose project name stay nemosyne [keep-nemosyne] (see deploy/docker-compose.yml).` Update the other mentions the same way.
- [ ] **Step 3: Verify.** `docker compose -f deploy/docker-compose.yml config` (with dummy `POSTGRES_PASSWORD` and `SESSION_SECRET`) shows project `nemosyne` and volume `nemosyne_gaia`, and the nginx volume mounts `kuronuma-proxy.conf`; YAML parses; every nginx `include` path matches a mounted file; `check-brand.sh` exits 0; linked paths exist.
- [ ] **Step 4: Commit** `refactor(deploy): rename the nginx include and update the docs for Kuronuma`.

### Task 6: Log, PR and CI

**Files:**
- Create: `agent-logs/claude/2026-10-31-HHMM-kuronuma-rename.md`.

- [ ] **Step 1: Write the log** (what changed, what was kept and why, verification, gaps).
- [ ] **Step 2: Open the PR** `refactor(brand): rename Nemosyne to Kuronuma` with the template: approvals ticked (deploy, CI, auth), the list of English and Thai copy lines for review, and a note that the address, Cloudflare, Google redirect URI and `PUBLIC_URL` are for the owner and are not in this PR.
- [ ] **Step 3: Wait for CI** (all four checks, including both image builds and `nginx -t`), then put the real result in the PR body.
