// The dashboard's client for the API. Every call goes through the web bridge (POST /api/web/call), which runs the
// same tools as the assistants for the signed-in user. Nothing here ever throws, and nothing a server or a proxy
// says reaches the page except the text a tool wrote for people (a 422).

export type Lang = "en" | "th";

export type WebResult<T = Record<string, unknown>> =
  | { ok: true; data: T }
  | { ok: false; kind: "signed-out" | "refused" | "unavailable"; message: string };

const generic: Record<Lang, string> = {
  en: "Something went wrong. Please try again.",
  th: "เกิดข้อผิดพลาดบางอย่าง โปรดลองอีกครั้ง",
};

const MAX_MESSAGE = 500;

type FetchLike = (url: string, init: RequestInit & { headers: Record<string, string> }) => Promise<Response>;

/** Sends one request to the API from our own page, with the headers the server requires. */
export async function webRequest<T = Record<string, unknown>>(
  method: string,
  path: string,
  body?: unknown,
  lang: Lang = "en",
  fetchImpl: FetchLike = fetch as unknown as FetchLike,
): Promise<WebResult<T>> {
  const unavailable: WebResult<T> = { ok: false, kind: "unavailable", message: generic[lang] };
  let res: Response;
  try {
    res = await fetchImpl(path, {
      method,
      credentials: "same-origin",
      cache: "no-store",
      headers: { "Content-Type": "application/json", "X-Kuronuma-Web": "1" },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
  } catch {
    return unavailable;
  }
  if (res.status === 401) return { ok: false, kind: "signed-out", message: "" };

  let json: unknown;
  try {
    json = await res.json();
  } catch {
    return unavailable; // an HTML error page from a proxy, or an empty body
  }
  if (res.status === 422) {
    const text = json && typeof json === "object" ? (json as { error?: unknown }).error : undefined;
    if (typeof text === "string" && text.trim() !== "") {
      return { ok: false, kind: "refused", message: text.trim().slice(0, MAX_MESSAGE) };
    }
    return unavailable;
  }
  if (res.status !== 200 || json === null || typeof json !== "object" || Array.isArray(json)) return unavailable;
  return { ok: true, data: json as T };
}

/** Runs one tool as the signed-in user. */
export async function callTool<T = Record<string, unknown>>(
  name: string,
  args: Record<string, unknown> = {},
  lang: Lang = "en",
  fetchImpl?: FetchLike,
): Promise<WebResult<T>> {
  const r = await webRequest<{ result?: unknown }>("POST", "/api/web/call", { tool: name, arguments: args }, lang, fetchImpl);
  if (!r.ok) return r;
  const result = r.data.result;
  if (result === null || typeof result !== "object" || Array.isArray(result)) {
    return { ok: false, kind: "unavailable", message: generic[lang] };
  }
  // The standing note about untrusted text is for assistants; the page never shows it.
  const { _note: _ignored, ...rest } = result as Record<string, unknown>;
  return { ok: true, data: rest as T };
}
