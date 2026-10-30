"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

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
    <main className="shell">
      <header className="brand">
        <h1>Connect your assistant</h1>
        <p className="tagline">One URL. Your assistant signs you in with Google and asks for your approval.</p>
      </header>

      <section className="card">
        <strong>Nemosyne MCP URL</strong>
        <div className="urlrow">
          <code data-testid="mcp-url">{mcpUrl || "Loading…"}</code>
          <button className="button ghost" onClick={copy} disabled={!mcpUrl}>
            {copied ? "Copied" : "Copy"}
          </button>
        </div>
        <p className="muted">
          Remote assistants need a public HTTPS address. If you run Nemosyne on your own computer, only apps on that
          computer can reach it.
        </p>
      </section>

      <h2>Claude</h2>
      <ol className="steps">
        <li>Open Settings, then Connectors, then add a custom connector.</li>
        <li>Paste the URL above and save.</li>
        <li>Choose Connect. Sign in with Google and review what Nemosyne asks to access, then approve.</li>
      </ol>
      <p className="muted">Menu names can differ between plans and app versions.</p>

      <h2>Claude Code</h2>
      <pre className="code">{`claude mcp add --transport http nemosyne ${mcpUrl || "<your-url>/mcp"}`}</pre>
      <p className="muted">Then run /mcp inside Claude Code and choose nemosyne to sign in.</p>

      <h2>Other MCP clients</h2>
      <p>
        Add a remote (Streamable HTTP) server with the URL above. Clients that support OAuth discovery register and
        sign in automatically.
      </p>

      <h2>Try it</h2>
      <ul className="steps">
        <li>Attach a bank slip photo and say: “Save this to Nemosyne.” Confirm the amount, payee and date it shows you.</li>
        <li>Ask: “How much did I spend on food this month?”</li>
      </ul>

      <p>
        <Link href="/">Back to dashboard</Link>
      </p>
    </main>
  );
}
