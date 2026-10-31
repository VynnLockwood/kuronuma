# Security policy

Nemosyne stores personal financial data, so security reports are welcome and taken seriously.

## Reporting a vulnerability

Please do not open a public issue. Use the **Report a vulnerability** button on the repository's Security tab (GitHub private vulnerability reporting). Include what you found, how to reproduce it, the impact you expect, and the commit or version you tested.

This is a volunteer project. We will acknowledge a report as soon as we can and keep you updated, but we cannot promise response times. Please give us a reasonable amount of time to fix a problem before you disclose it publicly.

## Scope

In scope: the code in this repository as shipped: the Go API (sign-in, OAuth server, MCP endpoint, modules), the web app, and the Docker and nginx configuration.

Out of scope:

- A self-hosted instance that is misconfigured (for example, left open to any Google account on purpose).
- Third-party services such as Google sign-in or the assistants that connect to Nemo.
- Social engineering, physical access, or a compromised user device.
- Denial of service by sheer traffic volume.

## Testing safely

Test only against your own instance or your own account. Do not read, change, or keep other people's data beyond what is needed to show the problem, and stop and report as soon as you can see another user's data.

## Supported versions

Only the latest commit on `main`. There are no versioned releases yet.

## If you run Nemo yourself

See the Security section of the [README](README.md): restrict who can sign in with `ALLOWED_EMAILS`, use strong secrets, terminate TLS in front of nginx, keep Postgres unpublished, and keep your images up to date.
