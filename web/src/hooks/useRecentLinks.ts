import { useCallback, useEffect, useState } from "react";
import type { ShortenResponse } from "../api/client";

export type RecentLink = ShortenResponse & {
  created_at: string;
};

const STORAGE_KEY = "nexuslink.recent.v1";
const LIMIT = 6;

function isRecentLink(value: unknown): value is RecentLink {
  if (!value || typeof value !== "object") {
    return false;
  }

  const item = value as Record<string, unknown>;
  return (
    typeof item.short_code === "string" &&
    typeof item.short_url === "string" &&
    typeof item.long_url === "string" &&
    typeof item.created_at === "string" &&
    !Number.isNaN(Date.parse(item.created_at)) &&
    (item.expires_at === undefined || typeof item.expires_at === "string")
  );
}

function readRecent(): RecentLink[] {
  if (typeof window === "undefined") {
    return [];
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      return [];
    }

    const parsed: unknown = JSON.parse(stored);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter(isRecentLink).slice(0, LIMIT);
  } catch {
    return [];
  }
}

function persist(items: RecentLink[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  } catch {
    /* Ignore quota and private mode failures. */
  }
}

export function useRecentLinks() {
  const [items, setItems] = useState<RecentLink[]>(readRecent);

  useEffect(() => {
    persist(items);
  }, [items]);

  useEffect(() => {
    function handleStorage(event: StorageEvent) {
      if (event.key === STORAGE_KEY || event.key === null) {
        setItems(readRecent());
      }
    }

    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, []);

  const add = useCallback((link: ShortenResponse) => {
    setItems((current) => [
      { ...link, created_at: new Date().toISOString() },
      ...current.filter((item) => item.short_code !== link.short_code),
    ].slice(0, LIMIT));
  }, []);

  const remove = useCallback((shortCode: string) => {
    setItems((current) => current.filter((item) => item.short_code !== shortCode));
  }, []);

  const clear = useCallback(() => {
    setItems([]);
  }, []);

  return { items, add, remove, clear };
}
