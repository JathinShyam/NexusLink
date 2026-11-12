import { useEffect, useState } from "react";
import type { ShortenRequest } from "../api/client";
import { useAuth } from "../auth/useAuth";
import { ManageDialog } from "../components/ManageDialog";
import { RecentLinks } from "../components/RecentLinks";
import { ResultPanel } from "../components/ResultPanel";
import { ShortenForm } from "../components/ShortenForm";
import { usePageTitle } from "../hooks/usePageTitle";
import { useRecentLinks } from "../hooks/useRecentLinks";
import { useShorten } from "../hooks/useShorten";
import { navigate, useLocation } from "../lib/navigation";

const MANAGE_SIGN_IN = `/signin?next=${encodeURIComponent("/?manage=1")}`;

export function HomePage() {
  const { token, user, isAuthenticated, clearSession } = useAuth();
  const { result, error, errorCode, create, reset, dismissError, isLoading } =
    useShorten(token);
  const { items: recentLinks, add: addRecentLink, remove: removeRecentLink } =
    useRecentLinks();
  const location = useLocation();

  const [manageOpen, setManageOpen] = useState(
    () => new URLSearchParams(location.search).get("manage") === "1",
  );
  const [focusSignal, setFocusSignal] = useState(0);

  usePageTitle();

  useEffect(() => {
    if (new URLSearchParams(location.search).get("manage") !== "1") {
      return;
    }
    setManageOpen(true);
    navigate("/", { replace: true });
  }, [location.search]);

  function focusComposer() {
    setFocusSignal((signal) => signal + 1);
  }

  function handleReset() {
    reset();
    focusComposer();
  }

  async function handleCreate(values: ShortenRequest) {
    const response = await create(values);
    if (response) {
      addRecentLink(response);
    }
    return response;
  }

  function handleDeleted(shortCode: string) {
    removeRecentLink(shortCode);
    if (result?.short_code === shortCode) {
      reset();
    }
  }

  function handleManage() {
    if (isAuthenticated) {
      setManageOpen(true);
    } else {
      navigate(MANAGE_SIGN_IN);
    }
  }

  useEffect(() => {
    if (errorCode === "UNAUTHORIZED" && isAuthenticated) {
      clearSession();
    }
  }, [errorCode, isAuthenticated, clearSession]);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (
        event.key !== "/" ||
        event.metaKey ||
        event.ctrlKey ||
        event.altKey ||
        document.querySelector("dialog[open]")
      ) {
        return;
      }

      const target = event.target as HTMLElement | null;
      if (
        target?.isContentEditable ||
        target?.tagName === "INPUT" ||
        target?.tagName === "TEXTAREA" ||
        target?.tagName === "SELECT"
      ) {
        return;
      }

      event.preventDefault();
      if (result) {
        reset();
      }
      focusComposer();
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [result, reset]);

  const previousLinks = recentLinks.filter(
    (link) => link.short_code !== result?.short_code,
  );

  return (
    <main id="main" className="app__main" tabIndex={-1}>
      <div className="desk shell">
        <div className="desk__intro">
          <h1 className="intro__title">
            Long links,
            <br />
            cut to size.
          </h1>
          <p className="intro__lede">
            Paste a URL. Get a short link you can copy, share, and own.
          </p>

          <p className="shortcut-hint">
            Press <kbd className="kbd">/</kbd> to focus the URL field
          </p>
        </div>

        <div className="desk__work">
          <div className="stage">
            {result ? (
              <ResultPanel result={result} onReset={handleReset} />
            ) : (
              <ShortenForm
                isLoading={isLoading}
                error={error}
                errorCode={errorCode}
                isAuthenticated={isAuthenticated}
                focusSignal={focusSignal}
                onSubmit={handleCreate}
                onDismissError={dismissError}
                onManage={handleManage}
              />
            )}
          </div>

          <dl className="facts">
            <div className="facts__item">
              <dt>Anonymous</dt>
              <dd>Shorten without signing up.</dd>
            </div>
            <div className="facts__item">
              <dt>Expiry</dt>
              <dd>Choose when a link stops working.</dd>
            </div>
            <div className="facts__item">
              <dt>Ownership</dt>
              <dd>Sign in to keep links you can manage.</dd>
            </div>
          </dl>

          {previousLinks.length > 0 ? (
            <RecentLinks items={previousLinks} onRemove={removeRecentLink} />
          ) : (
            <section className="recent recent--empty" aria-labelledby="recent-title">
              <div className="recent__head">
                <h2 id="recent-title" className="recent__title">
                  Recent links
                </h2>
                <p className="recent__hint">Kept in this browser</p>
              </div>
              <p className="recent__empty">
                Links you create will stack up here, ready to copy again.
              </p>
            </section>
          )}
        </div>
      </div>

      <ManageDialog
        open={manageOpen}
        token={token}
        user={user}
        isAuthenticated={isAuthenticated}
        onClose={() => setManageOpen(false)}
        onSignIn={() => {
          setManageOpen(false);
          navigate(MANAGE_SIGN_IN);
        }}
        onSessionExpired={clearSession}
        onDeleted={handleDeleted}
      />
    </main>
  );
}
