import { useRef, useState, type FormEvent } from "react";
import { useAuth } from "../auth/useAuth";
import { usePageTitle } from "../hooks/usePageTitle";
import { navigate, safeNextPath, useLocation } from "../lib/navigation";
import { AppLink } from "../lib/router";
import { AlertIcon } from "../components/Icons";
import { AuthLayout } from "./AuthLayout";

type Props = {
  mode: "signin" | "register";
};

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const PASSWORD_MIN = 8;
const PASSWORD_MAX = 72;

type FieldErrors = {
  email?: string;
  password?: string;
};

function validateFields(email: string, password: string, mode: Props["mode"]): FieldErrors {
  const errors: FieldErrors = {};
  const trimmedEmail = email.trim();

  if (!trimmedEmail) {
    errors.email = "Enter your email address.";
  } else if (!EMAIL_PATTERN.test(trimmedEmail)) {
    errors.email = "Enter a valid email, like you@example.com.";
  }

  if (!password) {
    errors.password = "Enter your password.";
  } else if (
    mode === "register" &&
    (password.length < PASSWORD_MIN || password.length > PASSWORD_MAX)
  ) {
    errors.password = "Use between 8 and 72 characters.";
  }

  return errors;
}

export function AuthPage({ mode }: Props) {
  const { status, user, isAuthenticated, pending, error, signIn, register, clearError } =
    useAuth();
  const location = useLocation();
  const next = safeNextPath(new URLSearchParams(location.search).get("next"));
  const isRegister = mode === "register";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const emailRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);

  usePageTitle(isRegister ? "Create account" : "Sign in");

  const points = isRegister
    ? [
        { term: "Ownership", copy: "Every link you make is tied to your account." },
        { term: "Cleanup", copy: "Delete an owned link whenever you want." },
        { term: "Simplicity", copy: "No inbox, no newsletter. Just your links." },
      ]
    : [
        { term: "Ownership", copy: "Links made while signed in belong to you." },
        { term: "Cleanup", copy: "Delete any owned link by its short code." },
        { term: "Privacy", copy: "Your session stays in this browser." },
      ];

  const intro = isRegister
    ? {
        title: (
          <>
            Take a
            <br />
            seat.
          </>
        ),
        lede: "Make an account to keep your short links together — and remove them whenever you want.",
      }
    : {
        title: (
          <>
            Back to
            <br />
            your desk.
          </>
        ),
        lede: "Your links are waiting. Sign in to create owned links and clean up the ones you no longer need.",
      };

  function handleEmailChange(value: string) {
    setEmail(value);
    if (fieldErrors.email) {
      setFieldErrors((current) => ({ ...current, email: undefined }));
    }
    if (error) {
      clearError();
    }
  }

  function handlePasswordChange(value: string) {
    setPassword(value);
    if (fieldErrors.password) {
      setFieldErrors((current) => ({ ...current, password: undefined }));
    }
    if (error) {
      clearError();
    }
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (pending) {
      return;
    }

    const errors = validateFields(email, password, mode);
    setFieldErrors(errors);

    if (errors.email || errors.password) {
      if (errors.email) {
        emailRef.current?.focus();
      } else {
        passwordRef.current?.focus();
      }
      return;
    }

    const result = isRegister
      ? await register(email.trim(), password, displayName)
      : await signIn(email.trim(), password);

    if (result.ok) {
      navigate(next, { replace: true });
    }
  }

  const card = (() => {
    if (status === "loading") {
      return (
        <div className="auth-card" aria-busy="true">
          <h2 className="auth-card__title">Checking your session…</h2>
          <p className="auth-card__lede">
            One moment while we confirm whether you’re already signed in.
          </p>
          <p className="auth-loading">
            <span className="spinner" aria-hidden="true" />
            <span>Checking session</span>
          </p>
        </div>
      );
    }

    if (isAuthenticated && user) {
      return (
        <div className="auth-card">
          <h2 className="auth-card__title">You’re already signed in.</h2>
          <p className="auth-card__lede">
            Continue as {user.display_name || user.email}. Nothing else to do here.
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

          <AppLink className="btn btn--primary auth-submit" to={next}>
            Continue to NexusLink
          </AppLink>
          <p className="auth-switch">
            Not you? <AppLink to="/signout">Sign out</AppLink>
          </p>
        </div>
      );
    }

    return (
      <div className="auth-card">
        <h2 className="auth-card__title">
          {isRegister ? "Create your account" : "Sign in"}
        </h2>
        <p className="auth-card__lede">
          {isRegister
            ? "Use an email and a password you’ll remember. That’s it."
            : "Enter the email and password you signed up with."}
        </p>

        <form className="auth-form" onSubmit={handleSubmit} noValidate>
          {error ? (
            <p className="form-error auth-error" role="alert">
              <AlertIcon className="icon" />
              <span>{error}</span>
            </p>
          ) : null}

          {isRegister ? (
            <div className="field">
              <label className="field__label" htmlFor="auth-name">
                Display name <span className="auth-optional">Optional</span>
              </label>
              <input
                id="auth-name"
                className="field__control"
                type="text"
                name="name"
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
                placeholder="Alex"
                autoComplete="name"
                maxLength={80}
                disabled={pending}
              />
            </div>
          ) : null}

          <div className="field">
            <label className="field__label" htmlFor="auth-email">
              Email
            </label>
            <input
              ref={emailRef}
              id="auth-email"
              className="field__control"
              type="email"
              name="email"
              value={email}
              onChange={(event) => handleEmailChange(event.target.value)}
              placeholder="you@example.com"
              autoComplete="email"
              inputMode="email"
              autoCapitalize="none"
              spellCheck={false}
              required
              disabled={pending}
              autoFocus={!isRegister}
              aria-invalid={Boolean(fieldErrors.email)}
              aria-describedby={fieldErrors.email ? "auth-email-error" : undefined}
            />
            {fieldErrors.email ? (
              <span id="auth-email-error" className="field__hint" data-tone="danger">
                {fieldErrors.email}
              </span>
            ) : null}
          </div>

          <div className="field">
            <label className="field__label" htmlFor="auth-password">
              Password
            </label>
            <div className="auth-password">
              <input
                ref={passwordRef}
                id="auth-password"
                className="field__control auth-password__input"
                type={showPassword ? "text" : "password"}
                name="password"
                value={password}
                onChange={(event) => handlePasswordChange(event.target.value)}
                placeholder={isRegister ? "At least 8 characters" : "Your password"}
                autoComplete={isRegister ? "new-password" : "current-password"}
                required
                disabled={pending}
                aria-invalid={Boolean(fieldErrors.password)}
                aria-describedby={
                  fieldErrors.password
                    ? "auth-password-error"
                    : isRegister
                      ? "auth-password-hint"
                      : undefined
                }
              />
              <button
                className="auth-password__toggle"
                type="button"
                onClick={() => setShowPassword((shown) => !shown)}
                aria-pressed={showPassword}
                aria-controls="auth-password"
                disabled={pending}
              >
                {showPassword ? "Hide" : "Show"}
              </button>
            </div>
            {fieldErrors.password ? (
              <span id="auth-password-error" className="field__hint" data-tone="danger">
                {fieldErrors.password}
              </span>
            ) : isRegister ? (
              <span id="auth-password-hint" className="field__hint">
                8–72 characters. A short passphrase works well.
              </span>
            ) : null}
          </div>

          <button
            className="btn btn--primary auth-submit"
            type="submit"
            disabled={pending}
            aria-busy={pending}
          >
            {pending ? (
              <>
                <span className="spinner" aria-hidden="true" />
                {isRegister ? "Creating account…" : "Signing in…"}
              </>
            ) : isRegister ? (
              "Create account"
            ) : (
              "Sign in"
            )}
          </button>
        </form>

        <p className="auth-switch">
          {isRegister ? (
            <>
              Already have an account?{" "}
              <AppLink to={`/signin${next === "/" ? "" : `?next=${encodeURIComponent(next)}`}`}>
                Sign in
              </AppLink>
            </>
          ) : (
            <>
              New here?{" "}
              <AppLink
                to={`/register${next === "/" ? "" : `?next=${encodeURIComponent(next)}`}`}
              >
                Create an account
              </AppLink>
            </>
          )}
        </p>
      </div>
    );
  })();

  return (
    <AuthLayout title={intro.title} lede={intro.lede} points={points}>
      {card}
    </AuthLayout>
  );
}
