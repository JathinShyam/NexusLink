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

export type AuthUser = {
  id: string;
  email: string;
  display_name: string;
  created_at: string;
};

export type AuthResponse = {
  token: string;
  expires_at: string;
  user: AuthUser;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type RegisterRequest = {
  email: string;
  password: string;
  display_name?: string;
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
const REQUEST_TIMEOUT = 20_000;

async function parseError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as ApiErrorBody;
    return new ApiError(res.status, body.code || "UNKNOWN", body.error || res.statusText);
  } catch {
    return new ApiError(res.status, "UNKNOWN", res.statusText || "request failed");
  }
}

async function request(url: string, init: RequestInit): Promise<Response> {
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT);

  try {
    return await fetch(url, {
      ...init,
      headers: {
        Accept: "application/json",
        ...init.headers,
      },
      signal: controller.signal,
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      throw new ApiError(0, "TIMEOUT", "request timed out");
    }
    throw error;
  } finally {
    window.clearTimeout(timer);
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

  const res = await request(`${baseURL}/api/v1/shorten`, {
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
  const res = await request(`${baseURL}/api/v1/links/${encodeURIComponent(shortCode)}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) {
    throw await parseError(res);
  }
}

export async function login(payload: LoginRequest): Promise<AuthResponse> {
  const res = await request(`${baseURL}/api/v1/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    throw await parseError(res);
  }

  return (await res.json()) as AuthResponse;
}

export async function register(payload: RegisterRequest): Promise<AuthResponse> {
  const res = await request(`${baseURL}/api/v1/auth/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    throw await parseError(res);
  }

  return (await res.json()) as AuthResponse;
}

export async function logout(token: string): Promise<void> {
  const res = await request(`${baseURL}/api/v1/auth/logout`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) {
    throw await parseError(res);
  }
}

export async function fetchMe(token: string): Promise<AuthUser> {
  const res = await request(`${baseURL}/api/v1/auth/me`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) {
    throw await parseError(res);
  }

  const body = (await res.json()) as { user: AuthUser };
  return body.user;
}

export async function checkReadiness(signal?: AbortSignal): Promise<boolean> {
  try {
    const res = await fetch(`${baseURL}/ready`, {
      headers: { Accept: "application/json" },
      cache: "no-store",
      signal,
    });
    return res.ok;
  } catch {
    return false;
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
        return "Your session expired. Sign in again.";
      case "INVALID_CREDENTIALS":
        return "Email or password is incorrect.";
      case "EMAIL_TAKEN":
        return "An account with that email already exists. Sign in instead.";
      case "WEAK_PASSWORD":
        return "Use a password between 8 and 72 characters.";
      case "INVALID_EMAIL":
        return "Enter a valid email address, like you@example.com.";
      case "FORBIDDEN":
        return err.message || "You cannot modify this link.";
      case "NOT_FOUND":
        return "Link not found or already deleted.";
      case "TIMEOUT":
        return "The request took too long. Check your connection and try again.";
      default:
        return err.message || "Something went wrong.";
    }
  }
  if (err instanceof TypeError) {
    return "Can’t reach the API. Check that the server is running, then try again.";
  }
  if (err instanceof Error) {
    return err.message;
  }
  return "Something went wrong.";
}
