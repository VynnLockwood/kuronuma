# Nemosyne

> Memory for your AI assistant. Named for the Greek goddess of memory (classically spelled *Mnemosyne*), mother of the Muses.

Nemosyne is an open-source personal data hub exposed to AI assistants over [MCP](https://modelcontextprotocol.io). Users keep their spending, tasks, routines and work data in one place, and their own assistant reads and writes it through scoped tools.

**Status:** design phase. No application code yet. See [docs/design.md](docs/design.md).

## First use case

Snap a bank slip, attach it in your assistant's chat, confirm the extracted amount, payee and date, and Nemosyne stores it. Ask for reports any time.

## Stack

| Layer | Choice |
|---|---|
| Frontend | Next.js (TypeScript) |
| Backend | Go |
| Database | PostgreSQL |
| Edge | nginx |
| Packaging | Docker / Docker Compose |
| Protocol | MCP over Streamable HTTP, OAuth for remote clients |

## The pantheon

| Name | Role |
|---|---|
| **Olympus** | Next.js web dashboard |
| **Propylaea** | nginx edge: TLS and routing |
| **Hermes** | MCP gateway between assistants and data |
| **Cerberus** | Auth and OAuth guard |
| **Hades** | Encrypted vault: field-level encryption and keys |
| **Gaia** | PostgreSQL storage |
| **Pythia** | Slip intake: validation, dedupe, review queue |
| **Plutus** | Spending module (first module) |
| **Atlas** | Tasks module |
| **Horae** | Routines module |
| **Hephaestus** | Work data module |
| **Moirai** | Background jobs and scheduling (later) |

## License

To be decided.
