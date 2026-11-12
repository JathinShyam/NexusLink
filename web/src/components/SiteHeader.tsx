import { useEffect, useId, useRef, useState, type KeyboardEvent } from "react";
import { useAuth } from "../auth/useAuth";
import type { Theme } from "../hooks/useTheme";
import { useLocation } from "../lib/navigation";
import { AppLink } from "../lib/router";
import {
  ChevronDownIcon,
  KeyIcon,
  LogOutIcon,
  MoonIcon,
  RouteMark,
  SunIcon,
  UserIcon,
} from "./Icons";
import "./SiteHeader.css";

type Props = {
  theme: Theme;
  onToggleTheme: () => void;
};

function accountName(displayName: string, email: string): string {
  return displayName.trim() || email;
}

function initialFor(displayName: string, email: string): string {
  const value = accountName(displayName, email).trim();
  return value ? value.charAt(0).toUpperCase() : "?";
}

export function SiteHeader({ theme, onToggleTheme }: Props) {
  const { status, user, isAuthenticated } = useAuth();
  const location = useLocation();
  const nextTheme = theme === "dark" ? "light" : "dark";
  const menuId = useId();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    setMenuOpen(false);
  }, [location.href]);

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    function handlePointerDown(event: PointerEvent) {
      const target = event.target as Node;
      if (triggerRef.current?.contains(target) || panelRef.current?.contains(target)) {
        return;
      }
      setMenuOpen(false);
    }

    document.addEventListener("pointerdown", handlePointerDown);
    return () => document.removeEventListener("pointerdown", handlePointerDown);
  }, [menuOpen]);

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    const frame = window.requestAnimationFrame(() => {
      panelRef.current?.querySelector<HTMLElement>('[role="menuitem"]')?.focus();
    });
    return () => window.cancelAnimationFrame(frame);
  }, [menuOpen]);

  function closeMenu(restoreFocus = false) {
    setMenuOpen(false);
    if (restoreFocus) {
      triggerRef.current?.focus();
    }
  }

  function handleTriggerKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (event.key === "ArrowDown" || event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      setMenuOpen(true);
    }
  }

  function handleMenuKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === "Escape") {
      event.preventDefault();
      closeMenu(true);
    }
  }

  const signInHref =
    location.pathname === "/" ||
    location.pathname === "/signin" ||
    location.pathname === "/register" ||
    location.pathname === "/signout"
      ? "/signin"
      : `/signin?next=${encodeURIComponent(location.href)}`;

  return (
    <header className="site-header">
      <div className="site-header__inner shell">
        <AppLink className="brand" to="/" aria-label="NexusLink home">
          <RouteMark className="brand__mark" />
          <span className="brand__name" translate="no">
            NexusLink
          </span>
        </AppLink>

        <div className="site-header__actions">
          {status === "loading" ? (
            <span className="site-header__auth-pending" aria-live="polite">
              <span className="sr-only">Checking your session…</span>
              <span className="site-header__auth-pending-bar" aria-hidden="true" />
            </span>
          ) : isAuthenticated && user ? (
            <div className="account">
              <button
                ref={triggerRef}
                className="account__trigger"
                type="button"
                aria-haspopup="menu"
                aria-expanded={menuOpen}
                aria-controls={menuOpen ? menuId : undefined}
                onClick={() => setMenuOpen((open) => !open)}
                onKeyDown={handleTriggerKeyDown}
              >
                <span className="account__avatar" aria-hidden="true">
                  {initialFor(user.display_name, user.email)}
                </span>
                <span className="account__name">
                  {accountName(user.display_name, user.email)}
                </span>
                <span className="sr-only">, account menu for {user.email}</span>
                <ChevronDownIcon className={menuOpen ? "icon is-open" : "icon"} />
              </button>

              {menuOpen ? (
                <div ref={panelRef} className="account__panel">
                  <p className="account__meta">
                    Signed in as <strong>{user.email}</strong>
                  </p>
                  <div
                    id={menuId}
                    role="menu"
                    aria-label="Account"
                    onKeyDown={handleMenuKeyDown}
                  >
                    <AppLink
                      className="account__item"
                      to="/?manage=1"
                      role="menuitem"
                      onClick={() => setMenuOpen(false)}
                    >
                      <KeyIcon className="icon" />
                      <span>Manage links</span>
                    </AppLink>
                    <AppLink
                      className="account__item"
                      to="/signout"
                      role="menuitem"
                      onClick={() => setMenuOpen(false)}
                    >
                      <LogOutIcon className="icon" />
                      <span>Sign out</span>
                    </AppLink>
                  </div>
                </div>
              ) : null}
            </div>
          ) : (
            <AppLink
              className="btn btn--quiet site-header__signin"
              to={signInHref}
              aria-label="Sign in"
            >
              <UserIcon className="icon" />
              <span>Sign in</span>
            </AppLink>
          )}

          <button
            className="site-header__theme"
            type="button"
            onClick={onToggleTheme}
            aria-label={`Switch to ${nextTheme} theme`}
            title={`Switch to ${nextTheme} theme`}
          >
            {theme === "dark" ? <SunIcon className="icon" /> : <MoonIcon className="icon" />}
          </button>
        </div>
      </div>
    </header>
  );
}
