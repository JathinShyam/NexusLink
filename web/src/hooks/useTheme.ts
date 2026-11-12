import { useCallback, useEffect, useState } from "react";
import { flushSync } from "react-dom";

export type Theme = "light" | "dark";

export type ThemeOrigin = {
  x: number;
  y: number;
};

const STORAGE_KEY = "nexuslink.theme";

type ViewTransitionHandle = {
  finished: Promise<void>;
};

type ViewTransitionDocument = Document & {
  startViewTransition?: (callback: () => void) => ViewTransitionHandle;
};

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

function prefersReducedMotion(): boolean {
  return window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? false;
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

  const toggleTheme = useCallback(
    (origin?: ThemeOrigin) => {
      const next: Theme = theme === "dark" ? "light" : "dark";

      const commit = () => {
        applyTheme(next);
        try {
          localStorage.setItem(STORAGE_KEY, next);
        } catch {
          /* ignore private mode */
        }
        flushSync(() => setTheme(next));
      };

      const doc = document as ViewTransitionDocument;
      if (prefersReducedMotion() || typeof doc.startViewTransition !== "function") {
        commit();
        return;
      }

      // The new palette opens from the toggle that was pressed.
      const root = document.documentElement;
      if (origin) {
        root.style.setProperty("--theme-x", `${origin.x}px`);
        root.style.setProperty("--theme-y", `${origin.y}px`);
      }

      const transition = doc.startViewTransition(commit);
      transition.finished
        .catch(() => undefined)
        .finally(() => {
          root.style.removeProperty("--theme-x");
          root.style.removeProperty("--theme-y");
        });
    },
    [theme],
  );

  return { theme, toggleTheme };
}
