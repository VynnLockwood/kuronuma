"use client";

import { useEffect, useState } from "react";
import { buttonVariants } from "@/components/ui/button";
import { fetchSession, signInUrl } from "@/lib/api";

// Several buttons can be on one page; share a single request.
let sessionCheck: Promise<boolean> | null = null;
const isSignedIn = () => (sessionCheck ??= fetchSession().then((s) => s.status === "signed-in"));

type Props = {
  signIn: string;
  openApp: string;
  className?: string;
};

// Shows "Open dashboard" when a session exists, otherwise "Sign in with Google".
// Starts as the sign-in link so the button never jumps size while the session loads.
export function AuthButton({ signIn, openApp, className = buttonVariants() }: Props) {
  const [signedIn, setSignedIn] = useState(false);

  useEffect(() => {
    let alive = true;
    isSignedIn().then((ok) => {
      if (alive) setSignedIn(ok);
    });
    return () => {
      alive = false;
    };
  }, []);

  return (
    <a className={className} href={signedIn ? "/dashboard" : signInUrl("/dashboard")}>
      {signedIn ? openApp : signIn}
    </a>
  );
}
