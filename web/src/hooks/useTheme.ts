import { useCallback, useEffect, useState } from "react";

export type Theme = "light" | "dark";

const STORAGE_KEY = "nexuslink.theme";

function readStoredTheme(): Theme | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === "dark" || stored === "light" ? stored : null;
  } catch {
    return null;
  }
}

function readTheme(): Theme {
  if (typeof document === "undefined") {
    return "light";
  }

  const current = document.documentElement.dataset.theme;
  if (current === "dark" || current === "light") {
    return current;
  }

  return window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme;

  const meta = document.querySelector('meta[name="theme-color"]');
  meta?.setAttribute("content", theme === "dark" ? "#0b0c0e" : "#f2f2ed");
}

export function useTheme() {
  const [theme, setTheme] = useState<Theme>(readTheme);

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");

    function handleSystemTheme(event: MediaQueryListEvent) {
      if (readStoredTheme()) {
        return;
      }
      setTheme(event.matches ? "dark" : "light");
    }

    function handleStorage(event: StorageEvent) {
      if (event.key !== STORAGE_KEY) {
        return;
      }
      if (event.newValue === "dark" || event.newValue === "light") {
        setTheme(event.newValue);
      } else {
        setTheme(media.matches ? "dark" : "light");
      }
    }

    media.addEventListener("change", handleSystemTheme);
    window.addEventListener("storage", handleStorage);

    return () => {
      media.removeEventListener("change", handleSystemTheme);
      window.removeEventListener("storage", handleStorage);
    };
  }, []);

  const toggleTheme = useCallback(() => {
    setTheme((current) => {
      const next: Theme = current === "dark" ? "light" : "dark";
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch {
        /* ignore private mode */
      }
      return next;
    });
  }, []);

  return { theme, toggleTheme };
}
