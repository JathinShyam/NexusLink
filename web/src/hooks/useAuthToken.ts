import { useCallback, useSyncExternalStore } from "react";

const STORAGE_KEY = "nexuslink.jwt";

let memoryToken = "";
const listeners = new Set<() => void>();

function emit() {
  for (const listener of listeners) {
    listener();
  }
}

function readToken(): string {
  if (typeof window === "undefined") {
    return memoryToken;
  }
  try {
    return localStorage.getItem(STORAGE_KEY) ?? "";
  } catch {
    return memoryToken;
  }
}

function writeToken(value: string) {
  memoryToken = value;
  try {
    if (value) {
      localStorage.setItem(STORAGE_KEY, value);
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  } catch {
    /* ignore quota / private mode */
  }
  emit();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function useAuthToken() {
  const token = useSyncExternalStore(subscribe, readToken, () => "");

  const setToken = useCallback((value: string) => {
    writeToken(value.trim());
  }, []);

  const clearToken = useCallback(() => {
    writeToken("");
  }, []);

  return { token, setToken, clearToken, hasToken: token.length > 0 };
}
