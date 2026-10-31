"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Icon } from "@/components/icons";
import { SiteHeader } from "@/components/SiteHeader";
import { authErrorMessages, fetchSession, signInUrl, signOut, type SessionState } from "@/lib/api";

const modules = [
  { god: "Plutus", area: "Spending", note: "Bank slips, totals and reports", live: true },
  { god: "Atlas", area: "Tasks and notes", note: "Tasks, to-dos and notes", live: true },
  { god: "Horae", area: "Routines", note: "Coming soon", live: false },
  { god: "Hephaestus", area: "Work data", note: "Coming soon", live: false },
];

export default function Dashboard() {
  const [session, setSession] = useState<SessionState | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);

  useEffect(() => {
    // Cerberus reports sign-in problems as ?auth_error=code. Show a message, then clean the URL.
    const code = new URLSearchParams(window.location.search).get("auth_error");
    if (code) {
      setAuthError(authErrorMessages[code] ?? authErrorMessages.server_error);
      window.history.replaceState(null, "", window.location.pathname);
    }
    fetchSession().then(setSession);
  }, []);

  async function handleSignOut() {
    await signOut();
    setSession({ status: "signed-out" });
  }

  return (
    <>
      <SiteHeader>
        <Link className="btn btn-ghost btn-sm" href="/connect">
          Connect
        </Link>
      </SiteHeader>

      <main className="page">
        {session === null && <p className="muted">Loading…</p>}

        {session?.status === "unreachable" && (
          <p className="notice" role="alert">
            Can’t reach Nemo right now. Please try again in a moment.
          </p>
        )}

        {session?.status === "signed-out" && (
          <section className="card auth-card">
            <img src="/brand/nemo-logo.png" width={96} height={96} alt="" />
            <h1>Sign in to Nemo</h1>
            <p className="muted">Use your Google account to connect your assistant to your own data.</p>
            {authError && (
              <p className="notice" role="alert">
                {authError}
              </p>
            )}
            <a className="btn btn-primary btn-lg" href={signInUrl("/dashboard")}>
              Continue with Google
            </a>
            <Link className="muted-link" href="/">
              Back to home
            </Link>
          </section>
        )}

        {session?.status === "signed-in" && (
          <>
            <section className="card who">
              {session.user.picture ? (
                // Google avatar; no-referrer keeps our URL out of Google's logs.
                // eslint-disable-next-line @next/next/no-img-element
                <img src={session.user.picture} alt="" width={48} height={48} referrerPolicy="no-referrer" />
              ) : (
                <span className="avatar-fallback" aria-hidden="true">
                  {(session.user.name || session.user.email).charAt(0).toUpperCase()}
                </span>
              )}
              <div className="who-text">
                <strong>{session.user.name || session.user.email}</strong>
                <span className="muted">{session.user.email}</span>
              </div>
              <button className="btn btn-ghost btn-sm" onClick={handleSignOut}>
                Sign out
              </button>
            </section>

            <section className="card connect-card">
              <span className="icon-tile">
                <Icon name="plug" />
              </span>
              <div>
                <h2>Connect your assistant</h2>
                <p className="muted">Add Nemo to Claude or any assistant that supports MCP connectors.</p>
              </div>
              <Link className="btn btn-primary" href="/connect">
                Connect
              </Link>
            </section>

            <h2 className="page-h2">Your modules</h2>
            <ul className="module-grid">
              {modules.map((m) => (
                <li key={m.god} className={`card module-card${m.live ? " live" : ""}`}>
                  <span className={`badge${m.live ? " badge-ok" : ""}`}>{m.live ? "Available" : "Coming soon"}</span>
                  <h3>{m.god}</h3>
                  <p className="module-area">{m.area}</p>
                  <p>{m.live ? m.note : ""}</p>
                </li>
              ))}
            </ul>
          </>
        )}
      </main>
    </>
  );
}
