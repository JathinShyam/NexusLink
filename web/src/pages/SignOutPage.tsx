import { useAuth } from "../auth/useAuth";
import { CheckIcon, LogOutIcon } from "../components/Icons";
import { usePageTitle } from "../hooks/usePageTitle";
import { AppLink } from "../lib/router";
import { AuthLayout } from "./AuthLayout";

export function SignOutPage() {
  const { status, user, isAuthenticated, pending, notice, signOut } = useAuth();
  usePageTitle("Sign out");

  const points = [
    { term: "Links stay", copy: "Signing out never deletes a short link." },
    { term: "Revoked", copy: "The session is ended on the server, not just here." },
    { term: "Return", copy: "Sign back in whenever you need the desk again." },
  ];

  let card;

  if (status === "loading") {
    card = (
      <div className="auth-card" aria-busy="true">
        <h2 className="auth-card__title">Checking your session…</h2>
        <p className="auth-card__lede">
          One moment while we confirm who is signed in on this browser.
        </p>
        <p className="auth-loading">
          <span className="spinner" aria-hidden="true" />
          <span>Checking session</span>
        </p>
      </div>
    );
  } else if (isAuthenticated && user) {
    card = (
      <div className="auth-card">
        <span className="auth-badge" aria-hidden="true">
          <LogOutIcon className="icon" />
        </span>
        <h2 className="auth-card__title">Sign out of this browser?</h2>
        <p className="auth-card__lede">
          Your links stay online. This session ends on the server, and you’ll need
          your password to sign in again.
        </p>

        <div className="person auth-person">
          <span className="person__avatar" aria-hidden="true">
            {(user.display_name || user.email).charAt(0).toUpperCase()}
          </span>
          <span className="person__body">
            <span className="person__name">{user.display_name || user.email}</span>
            <span className="person__email">{user.email}</span>
          </span>
        </div>

        <div className="auth-actions">
          <button
            className="btn btn--primary auth-submit"
            type="button"
            onClick={() => void signOut()}
            disabled={pending}
            aria-busy={pending}
          >
            {pending ? (
              <>
                <span className="spinner" aria-hidden="true" />
                Signing out…
              </>
            ) : (
              "Sign out"
            )}
          </button>
          <AppLink className="btn btn--ghost auth-submit" to="/">
            Stay signed in
          </AppLink>
        </div>

        <p className="auth-note">
          <span className="auth-note__dot" aria-hidden="true" />
          Your short links stay live after you sign out.
        </p>
      </div>
    );
  } else {
    card = (
      <div className="auth-card">
        <span className="auth-badge" data-tone="success" aria-hidden="true">
          <CheckIcon className="icon" />
        </span>
        <h2 className="auth-card__title">
          {notice === "signed-out" ? "You’re signed out." : "No active session."}
        </h2>
        <p className="auth-card__lede">
          {notice === "signed-out"
            ? "This browser no longer holds your session. Your links are still online and ready when you return."
            : "You’re browsing as a guest. Sign in to create owned links and manage the ones you already made."}
        </p>

        <div className="auth-actions">
          <AppLink className="btn btn--primary auth-submit" to="/signin">
            Sign in
          </AppLink>
          <AppLink className="btn btn--ghost auth-submit" to="/">
            Back to NexusLink
          </AppLink>
        </div>
      </div>
    );
  }

  return (
    <AuthLayout
      title={
        <>
          Step
          <br />
          away.
        </>
      }
      lede="End this session when you’re done. Your links keep their place, and you can pick up where you left off later."
      points={points}
    >
      {card}
    </AuthLayout>
  );
}
