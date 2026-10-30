"use client";

import { useEffect, useState } from "react";
import {
  authErrorMessages,
  fetchSession,
  signInUrl,
  signOut,
  type SessionState,
} from "@/lib/api";

const modules = [
  { god: "Plutus", area: "Spending", note: "Bank slips and reports" },
  { god: "Atlas", area: "Tasks", note: "Coming later" },
  { god: "Horae", area: "Routines", note: "Coming later" },
  { god: "Hephaestus", area: "Work data", note: "Coming later" },
];

export default function Home() {
  const [session, setSession] = useState<SessionState | null>(null);
  const [authError, setAuthError] = useState<string | null>(null);

  useEffect(() => {
    // Read the error code Cerberus puts on the redirect, then drop it from the URL.
    const params = new URLSearchParams(window.location.search);
    const code = params.get("auth_error");
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
    <main className="shell">
      <header className="brand">
        <h1>Nemosyne</h1>
        <p className="tagline">Memory for your AI assistant</p>
      </header>

      {session === null && <p className="muted">Loading…</p>}

      {session?.status === "unreachable" && (
        <p className="notice" role="alert">
          Can’t reach the Nemosyne API right now. Please try again in a moment.
        </p>
      )}

      {session?.status === "signed-out" && (
        <section className="card">
          {authError && (
            <p className="notice" role="alert">
              {authError}
            </p>
          )}
          <p>Sign in to connect your assistant to your own data.</p>
          <a className="button" href={signInUrl("/")}>
            Continue with Google
          </a>
        </section>
      )}

      {session?.status === "signed-in" && (
        <>
          <section className="card who">
            {session.user.picture && (
              // Google avatar; no-referrer keeps our URL out of Google's logs.
              // eslint-disable-next-line @next/next/no-img-element
              <img src={session.user.picture} alt="" width={40} height={40} referrerPolicy="no-referrer" />
            )}
            <div>
              <strong>{session.user.name || session.user.email}</strong>
              <div className="muted">{session.user.email}</div>
            </div>
            <button className="button ghost" onClick={handleSignOut}>
              Sign out
            </button>
          </section>

          <h2>Your pantheon</h2>
          <ul className="grid">
            {modules.map((m) => (
              <li key={m.god} className="card module">
                <strong>{m.god}</strong>
                <span>{m.area}</span>
                <span className="muted">{m.note}</span>
              </li>
            ))}
          </ul>
        </>
      )}
    </main>
  );
}
