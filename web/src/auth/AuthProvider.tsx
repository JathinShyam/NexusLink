import {
  useCallback,
  useEffect,
  useMemo,
  useReducer,
  type ReactNode,
} from "react";
import {
  ApiError,
  fetchMe,
  friendlyError,
  login as apiLogin,
  logout as apiLogout,
  register as apiRegister,
  type AuthUser,
} from "../api/client";
import { useAuthToken } from "../hooks/useAuthToken";
import { getTokenExpiry } from "../lib/jwt";
import {
  AuthContext,
  type AuthContextValue,
  type AuthResult,
  type AuthStatus,
} from "./context";

const USER_STORAGE_KEY = "nexuslink.user.v1";
const LOGOUT_TIMEOUT = 3_000;

type AuthState = {
  status: AuthStatus;
  user: AuthUser | null;
  pending: boolean;
  error: string | null;
  notice: "signed-out" | null;
};

type Action =
  | { type: "checking" }
  | { type: "restore"; user: AuthUser }
  | { type: "pending" }
  | { type: "success"; user: AuthUser }
  | { type: "failure"; error: string }
  | { type: "anonymous" }
  | { type: "signed-out" }
  | { type: "clear-error" };

const initialState: AuthState = {
  status: "loading",
  user: null,
  pending: false,
  error: null,
  notice: null,
};

function reducer(state: AuthState, action: Action): AuthState {
  switch (action.type) {
    case "checking":
      return state.status === "authenticated"
        ? { ...state }
        : { ...state, status: "loading", user: null, pending: false };
    case "restore":
      return {
        ...state,
        status: "authenticated",
        user: action.user,
        pending: false,
        notice: null,
      };
    case "pending":
      return { ...state, pending: true, error: null };
    case "success":
      return {
        ...state,
        status: "authenticated",
        user: action.user,
        pending: false,
        error: null,
        notice: null,
      };
    case "failure":
      return { ...state, pending: false, error: action.error };
    case "anonymous":
      return {
        ...state,
        status: "anonymous",
        user: null,
        pending: false,
        error: null,
      };
    case "signed-out":
      return {
        ...state,
        status: "anonymous",
        user: null,
        pending: false,
        error: null,
        notice: "signed-out",
      };
    case "clear-error":
      return state.error ? { ...state, error: null } : state;
    default:
      return state;
  }
}

type StoredUser = {
  token: string;
  user: AuthUser;
};

function isAuthUser(value: unknown): value is AuthUser {
  if (!value || typeof value !== "object") {
    return false;
  }
  const user = value as Record<string, unknown>;
  return (
    typeof user.id === "string" &&
    typeof user.email === "string" &&
    typeof user.display_name === "string" &&
    typeof user.created_at === "string"
  );
}

function readStoredUser(token: string): AuthUser | null {
  try {
    const raw = localStorage.getItem(USER_STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed = JSON.parse(raw) as Partial<StoredUser>;
    if (parsed.token !== token || !isAuthUser(parsed.user)) {
      return null;
    }
    return parsed.user;
  } catch {
    return null;
  }
}

function writeStoredUser(token: string, user: AuthUser | null) {
  try {
    if (!token || !user) {
      localStorage.removeItem(USER_STORAGE_KEY);
      return;
    }
    const stored: StoredUser = { token, user };
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(stored));
  } catch {
    /* ignore quota and private mode failures */
  }
}

function isSessionRejection(error: unknown): boolean {
  return error instanceof ApiError && (error.status === 401 || error.status === 403);
}

function logoutWithDeadline(token: string): Promise<void> {
  return new Promise((resolve) => {
    const timer = window.setTimeout(resolve, LOGOUT_TIMEOUT);
    void apiLogout(token)
      .catch(() => undefined)
      .finally(() => {
        window.clearTimeout(timer);
        resolve();
      });
  });
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const { token, setToken } = useAuthToken();
  const [state, dispatch] = useReducer(reducer, initialState);

  useEffect(() => {
    if (!token) {
      writeStoredUser("", null);
      dispatch({ type: "anonymous" });
      return;
    }

    const expiry = getTokenExpiry(token);
    if (expiry && expiry.getTime() <= Date.now()) {
      writeStoredUser("", null);
      setToken("");
      dispatch({ type: "anonymous" });
      return;
    }

    const cached = readStoredUser(token);
    if (cached) {
      dispatch({ type: "restore", user: cached });
    } else {
      dispatch({ type: "checking" });
    }

    let cancelled = false;
    void fetchMe(token)
      .then((user) => {
        if (cancelled) {
          return;
        }
        writeStoredUser(token, user);
        dispatch({ type: "restore", user });
      })
      .catch((error: unknown) => {
        if (cancelled) {
          return;
        }
        if (isSessionRejection(error)) {
          writeStoredUser("", null);
          setToken("");
          dispatch({ type: "anonymous" });
          return;
        }
        if (!cached) {
          setToken("");
          dispatch({ type: "anonymous" });
        }
      });

    return () => {
      cancelled = true;
    };
  }, [token, setToken]);

  useEffect(() => {
    if (!token) {
      return;
    }

    const expiry = getTokenExpiry(token);
    if (!expiry) {
      return;
    }

    const delay = expiry.getTime() - Date.now();
    if (delay <= 0) {
      writeStoredUser("", null);
      setToken("");
      return;
    }

    const timer = window.setTimeout(() => {
      writeStoredUser("", null);
      setToken("");
    }, Math.min(delay, 2_147_483_647));

    return () => window.clearTimeout(timer);
  }, [token, setToken]);

  const signIn = useCallback(
    async (email: string, password: string): Promise<AuthResult> => {
      dispatch({ type: "pending" });
      try {
        const response = await apiLogin({ email, password });
        writeStoredUser(response.token, response.user);
        setToken(response.token);
        dispatch({ type: "success", user: response.user });
        return { ok: true, user: response.user };
      } catch (error) {
        dispatch({ type: "failure", error: friendlyError(error) });
        return { ok: false };
      }
    },
    [setToken],
  );

  const register = useCallback(
    async (email: string, password: string, displayName: string): Promise<AuthResult> => {
      dispatch({ type: "pending" });
      try {
        const response = await apiRegister({
          email,
          password,
          display_name: displayName.trim() || undefined,
        });
        writeStoredUser(response.token, response.user);
        setToken(response.token);
        dispatch({ type: "success", user: response.user });
        return { ok: true, user: response.user };
      } catch (error) {
        dispatch({ type: "failure", error: friendlyError(error) });
        return { ok: false };
      }
    },
    [setToken],
  );

  const clearSession = useCallback(() => {
    writeStoredUser("", null);
    setToken("");
    dispatch({ type: "anonymous" });
  }, [setToken]);

  const signOut = useCallback(async () => {
    dispatch({ type: "pending" });
    if (token) {
      await logoutWithDeadline(token);
    }
    writeStoredUser("", null);
    setToken("");
    dispatch({ type: "signed-out" });
  }, [token, setToken]);

  const clearError = useCallback(() => {
    dispatch({ type: "clear-error" });
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      status: state.status,
      user: state.user,
      token,
      pending: state.pending,
      error: state.error,
      notice: state.notice,
      isAuthenticated: state.status === "authenticated" && Boolean(state.user),
      signIn,
      register,
      signOut,
      clearSession,
      clearError,
    }),
    [
      state.status,
      state.user,
      state.pending,
      state.error,
      state.notice,
      token,
      signIn,
      register,
      signOut,
      clearSession,
      clearError,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
