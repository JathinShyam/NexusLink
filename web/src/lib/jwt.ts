export function getTokenExpiry(token: string): Date | null {
  const payload = token.split(".")[1];

  if (!payload) {
    return null;
  }

  try {
    const normalized = payload.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    const decoded = JSON.parse(atob(padded)) as { exp?: unknown };

    if (typeof decoded.exp !== "number") {
      return null;
    }

    const expiry = new Date(decoded.exp * 1000);
    return Number.isNaN(expiry.getTime()) ? null : expiry;
  } catch {
    return null;
  }
}
