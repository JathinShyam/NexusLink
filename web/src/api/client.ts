export type ShortenRequest = {
  long_url: string;
  custom_alias?: string;
  expires_in_seconds?: number;
};

export type ShortenResponse = {
  short_code: string;
  short_url: string;
  long_url: string;
  expires_at?: string;
};

export type ApiErrorBody = {
  error: string;
  code: string;
};

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

const baseURL = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, "") ?? "";

async function parseError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as ApiErrorBody;
    return new ApiError(res.status, body.code || "UNKNOWN", body.error || res.statusText);
  } catch {
    return new ApiError(res.status, "UNKNOWN", res.statusText || "request failed");
  }
}

export async function shorten(
  payload: ShortenRequest,
  token?: string | null,
): Promise<ShortenResponse> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${baseURL}/api/v1/shorten`, {
    method: "POST",
    headers,
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    throw await parseError(res);
  }

  return (await res.json()) as ShortenResponse;
}

export async function deleteLink(shortCode: string, token: string): Promise<void> {
  const res = await fetch(`${baseURL}/api/v1/links/${encodeURIComponent(shortCode)}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) {
    throw await parseError(res);
  }
}

export function friendlyError(err: unknown): string {
  if (err instanceof ApiError) {
    switch (err.code) {
      case "INVALID_URL":
        return "That URL is not allowed. Use http or https (no localhost/private hosts).";
      case "INVALID_ALIAS":
        return "Alias must be 3–16 characters: letters, numbers, hyphens, or underscores.";
      case "ALIAS_TAKEN":
        return "That custom alias is already taken.";
      case "RATE_LIMITED":
        return "Too many requests. Wait a moment and try again.";
      case "UNAUTHORIZED":
        return "Valid JWT required. Run make gentoken and paste the token.";
      case "FORBIDDEN":
        return err.message || "You cannot modify this link.";
      case "NOT_FOUND":
        return "Link not found or already deleted.";
      default:
        return err.message || "Something went wrong.";
    }
  }
  if (err instanceof Error) {
    return err.message;
  }
  return "Something went wrong.";
}
