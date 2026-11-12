import { useSyncExternalStore } from "react";

const NAVIGATION_EVENT = "nexuslink:navigate";

function subscribe(listener: () => void) {
  window.addEventListener("popstate", listener);
  window.addEventListener(NAVIGATION_EVENT, listener);
  return () => {
    window.removeEventListener("popstate", listener);
    window.removeEventListener(NAVIGATION_EVENT, listener);
  };
}

function getSnapshot() {
  return `${window.location.pathname}${window.location.search}`;
}

function getServerSnapshot() {
  return "/";
}

export type Location = {
  pathname: string;
  search: string;
  href: string;
};

export function useLocation(): Location {
  const href = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const url = new URL(href, "http://nexuslink.local");
  return { pathname: url.pathname, search: url.search, href };
}

export function navigate(to: string, options: { replace?: boolean } = {}) {
  const target = new URL(to, window.location.origin);

  if (target.origin !== window.location.origin) {
    window.location.assign(target.href);
    return;
  }

  const next = `${target.pathname}${target.search}${target.hash}`;
  const current = `${window.location.pathname}${window.location.search}${window.location.hash}`;

  if (next === current) {
    return;
  }

  if (options.replace) {
    window.history.replaceState(null, "", next);
  } else {
    window.history.pushState(null, "", next);
  }

  window.dispatchEvent(new Event(NAVIGATION_EVENT));
  window.scrollTo(0, 0);
}

// Only allow same-origin app paths as post-login destinations. Parsing with a
// fixed base catches paths such as "/\evil.com" or "/\t/evil.com", which the
// URL parser would otherwise normalize into an off-site protocol-relative URL.
const NEXT_BASE_URL = "http://nexuslink.local";

export function safeNextPath(raw: string | null): string {
  if (!raw || !raw.startsWith("/") || raw.startsWith("//")) {
    return "/";
  }

  let target: URL;
  try {
    target = new URL(raw, NEXT_BASE_URL);
  } catch {
    return "/";
  }

  if (target.origin !== NEXT_BASE_URL) {
    return "/";
  }

  if (
    target.pathname === "/signin" ||
    target.pathname === "/register" ||
    target.pathname === "/signout"
  ) {
    return "/";
  }

  return `${target.pathname}${target.search}${target.hash}`;
}
