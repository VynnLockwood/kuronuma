// Hand-written for now; replaced by a client generated from api/openapi.yaml later.

export type User = {
  id: string;
  email: string;
  name: string;
  picture: string;
};

export type SessionState =
  | { status: "signed-out" }
  | { status: "signed-in"; user: User }
  | { status: "unreachable" };

export async function fetchSession(): Promise<SessionState> {
  try {
    const res = await fetch("/api/auth/me", { credentials: "same-origin", cache: "no-store" });
    if (res.status === 401) return { status: "signed-out" };
    if (!res.ok) return { status: "unreachable" };
    return { status: "signed-in", user: (await res.json()) as User };
  } catch {
    return { status: "unreachable" };
  }
}

export async function signOut(): Promise<void> {
  await fetch("/api/auth/logout", { method: "POST", credentials: "same-origin" });
}

export const signInUrl = (next = "/dashboard") =>
  `/api/auth/google/login?next=${encodeURIComponent(next)}`;

// Short codes set by the API on /?auth_error=...
export const authErrorMessages: Record<string, string> = {
  access_denied: "Sign-in was cancelled.",
  invalid_flow: "That sign-in attempt expired or was invalid. Please try again.",
  exchange_failed: "Google could not confirm your sign-in. Please try again.",
  email_unverified: "Your Google account email is not verified.",
  not_allowed: "This account is not on the access list yet.",
  server_error: "Something went wrong on our side. Please try again.",
};
