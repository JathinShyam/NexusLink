const HAS_SCHEME = /^[a-z][a-z0-9+.-]*:/i;

export type UrlCheck = { ok: true; value: string } | { ok: false; error: string };

export function prepareUrl(raw: string): UrlCheck {
  const trimmed = raw.trim();

  if (!trimmed) {
    return { ok: false, error: "Paste a URL to shorten." };
  }

  const candidate = HAS_SCHEME.test(trimmed) ? trimmed : `https://${trimmed}`;

  let parsed: URL;
  try {
    parsed = new URL(candidate);
  } catch {
    return {
      ok: false,
      error: "Enter a complete URL, like https://example.com/page.",
    };
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return { ok: false, error: "Use an http or https link." };
  }

  if (!parsed.hostname || parsed.hostname.includes("%")) {
    return {
      ok: false,
      error: "Enter a complete URL, like https://example.com/page.",
    };
  }

  return { ok: true, value: parsed.toString() };
}

export function displayUrl(raw: string): string {
  try {
    const url = new URL(raw);
    const path = url.pathname === "/" ? "" : url.pathname;
    return `${url.hostname}${path}${url.search}`;
  } catch {
    return raw;
  }
}
