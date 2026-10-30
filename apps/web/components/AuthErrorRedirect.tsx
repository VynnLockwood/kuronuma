"use client";

import { useEffect } from "react";

// The API reports sign-in problems by redirecting to /?auth_error=code.
// The homepage is now marketing, so forward those to the dashboard, which shows the message.
export function AuthErrorRedirect() {
  useEffect(() => {
    const code = new URLSearchParams(window.location.search).get("auth_error");
    if (code) window.location.replace(`/dashboard?auth_error=${encodeURIComponent(code)}`);
  }, []);
  return null;
}
