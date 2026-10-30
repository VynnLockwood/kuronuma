"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Icon } from "@/components/icons";
import { SiteHeader } from "@/components/SiteHeader";

export default function Connect() {
  // The MCP endpoint lives on the same origin as this page (nginx routes /mcp to the API).
  const [origin, setOrigin] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    setOrigin(window.location.origin);
  }, []);

  const mcpUrl = origin ? `${origin}/mcp` : "";

  async function copy() {
    try {
      await navigator.clipboard.writeText(mcpUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard can be unavailable (insecure context); the URL is selectable on the page.
    }
  }

  return (
    <>
      <SiteHeader>
        <Link className="btn btn-ghost btn-sm" href="/dashboard">
          Dashboard
        </Link>
      </SiteHeader>

      <main className="page">
        <h1 className="page-h1">Connect your assistant</h1>
        <p className="lead">One URL. Your assistant signs you in with Google and asks for your approval.</p>

        <section className="card url-card">
          <strong>Your Nemo MCP URL</strong>
          <div className="urlrow">
            <code data-testid="mcp-url">{mcpUrl || "Loading…"}</code>
            <button className="btn btn-primary btn-sm" onClick={copy} disabled={!mcpUrl}>
              {copied ? "Copied" : "Copy"}
            </button>
          </div>
          <p className="muted">
            Remote assistants need a public HTTPS address. If you run Nemo on your own computer, only apps on that
            computer can reach it.
          </p>
        </section>

        <div className="guide-grid">
          <section className="card guide">
            <span className="icon-tile">
              <Icon name="chat" />
            </span>
            <h2>Claude</h2>
            <ol className="steps">
              <li>Open Settings, then Connectors, then add a custom connector.</li>
              <li>Paste the URL above and save.</li>
              <li>Choose Connect. Sign in with Google, review what Nemo asks to access, then approve.</li>
            </ol>
            <p className="muted">Menu names can differ between plans and app versions.</p>
          </section>

          <section className="card guide">
            <span className="icon-tile">
              <Icon name="code" />
            </span>
            <h2>Claude Code</h2>
            <pre className="code">{`claude mcp add --transport http nemosyne ${mcpUrl || "<your-url>/mcp"}`}</pre>
            <p className="muted">Then run /mcp inside Claude Code and choose nemosyne to sign in.</p>
          </section>

          <section className="card guide">
            <span className="icon-tile">
              <Icon name="plug" />
            </span>
            <h2>Other MCP clients</h2>
            <p>
              Add a remote (Streamable HTTP) server with the URL above. Clients that support OAuth discovery register
              and sign in automatically.
            </p>
          </section>

          <section className="card guide">
            <span className="icon-tile">
              <Icon name="sparkle" />
            </span>
            <h2>Try it</h2>
            <ul className="steps">
              <li>Attach a bank slip photo and say: “Save this to Nemo.” Confirm the amount, payee and date it shows you.</li>
              <li>Ask: “How much did I spend on food this month?”</li>
            </ul>
          </section>
        </div>
      </main>
    </>
  );
}
