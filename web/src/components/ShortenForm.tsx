import { useEffect, useRef, useState, type FormEvent } from "react";
import type { ShortenRequest, ShortenResponse } from "../api/client";
import { prepareUrl } from "../lib/url";
import { AlertIcon, LinkIcon, PlusIcon } from "./Icons";
import "./ShortenForm.css";

type Props = {
  isLoading: boolean;
  error: string | null;
  errorCode: string | null;
  isAuthenticated: boolean;
  focusSignal: number;
  onSubmit: (values: ShortenRequest) => Promise<ShortenResponse | null>;
  onDismissError: () => void;
  onManage: () => void;
};

const EXPIRY_OPTIONS = [
  { value: "0", label: "Never" },
  { value: "1", label: "1 hour" },
  { value: "24", label: "24 hours" },
  { value: "168", label: "7 days" },
  { value: "720", label: "30 days" },
  { value: "custom", label: "Custom…" },
];

const ALIAS_PATTERN = /^[A-Za-z0-9_-]{3,16}$/;
const CUSTOM_HOURS_MAX = 8_760;

export function ShortenForm({
  isLoading,
  error,
  errorCode,
  isAuthenticated,
  focusSignal,
  onSubmit,
  onDismissError,
  onManage,
}: Props) {
  const [longUrl, setLongUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [expiryHours, setExpiryHours] = useState("0");
  const [customHours, setCustomHours] = useState("");
  const [showOptions, setShowOptions] = useState(false);
  const [urlError, setUrlError] = useState<string | null>(null);
  const [customError, setCustomError] = useState<string | null>(null);
  const [aliasTouched, setAliasTouched] = useState(false);

  const inputRef = useRef<HTMLInputElement>(null);
  const aliasRef = useRef<HTMLInputElement>(null);
  const customExpiryRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (focusSignal > 0) {
      inputRef.current?.focus();
    }
  }, [focusSignal]);

  useEffect(() => {
    if (expiryHours === "custom") {
      customExpiryRef.current?.focus();
    }
  }, [expiryHours]);

  useEffect(() => {
    if (errorCode === "INVALID_URL") {
      inputRef.current?.focus();
    }
  }, [errorCode]);

  useEffect(() => {
    if (errorCode === "INVALID_ALIAS" || errorCode === "ALIAS_TAKEN") {
      setShowOptions(true);
    }
  }, [errorCode]);

  useEffect(() => {
    if (
      showOptions &&
      (errorCode === "INVALID_ALIAS" || errorCode === "ALIAS_TAKEN")
    ) {
      aliasRef.current?.focus();
    }
  }, [errorCode, showOptions]);

  const trimmedAlias = alias.trim();
  const aliasInvalid = trimmedAlias.length > 0 && !ALIAS_PATTERN.test(trimmedAlias);
  const aliasApiError = errorCode === "INVALID_ALIAS" || errorCode === "ALIAS_TAKEN";
  const showAliasError = (aliasInvalid && aliasTouched) || aliasApiError;
  const urlFieldInvalid = Boolean(urlError) || errorCode === "INVALID_URL";
  const hasOptionValues = trimmedAlias.length > 0 || expiryHours !== "0";
  const visibleError = urlError ?? error;

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    setUrlError(null);
    setCustomError(null);
    setAliasTouched(true);

    const prepared = prepareUrl(longUrl);
    if (!prepared.ok) {
      setUrlError(prepared.error);
      inputRef.current?.focus();
      return;
    }

    if (aliasInvalid) {
      if (showOptions) {
        aliasRef.current?.focus();
      } else {
        setShowOptions(true);
        window.requestAnimationFrame(() => aliasRef.current?.focus());
      }
      return;
    }

    let hours = Number(expiryHours);
    if (expiryHours === "custom") {
      const parsed = Number(customHours);
      if (!Number.isInteger(parsed) || parsed < 1 || parsed > CUSTOM_HOURS_MAX) {
        setCustomError("Enter a whole number from 1 to 8,760 hours.");
        if (showOptions) {
          customExpiryRef.current?.focus();
        } else {
          setShowOptions(true);
          window.requestAnimationFrame(() => customExpiryRef.current?.focus());
        }
        return;
      }
      hours = parsed;
    }

    setLongUrl(prepared.value);
    await onSubmit({
      long_url: prepared.value,
      custom_alias: trimmedAlias || undefined,
      expires_in_seconds: hours > 0 ? Math.round(hours * 3600) : undefined,
    });
  }

  function dismissApiError() {
    if (errorCode) {
      onDismissError();
    }
  }

  function handleUrlChange(value: string) {
    setLongUrl(value);
    if (urlError) {
      setUrlError(null);
    }
    dismissApiError();
  }

  function handleAliasChange(value: string) {
    setAlias(value);
    dismissApiError();
  }

  function handleExpiryChange(value: string) {
    setExpiryHours(value);
    dismissApiError();
  }

  function handleCustomHoursChange(value: string) {
    setCustomHours(value);
    if (customError) {
      setCustomError(null);
    }
    dismissApiError();
  }

  return (
    <form className="composer" onSubmit={handleSubmit} noValidate aria-busy={isLoading}>
      <div className="composer__surface">
        <div className="composer__bar">
          <LinkIcon className="composer__leading-icon" />
          <label className="sr-only" htmlFor="long-url">
            Long URL
          </label>
          <input
            ref={inputRef}
            id="long-url"
            className="composer__input"
            type="text"
            name="long_url"
            required
            placeholder="https://example.com/your-link…"
            value={longUrl}
            onChange={(event) => handleUrlChange(event.target.value)}
            autoComplete="url"
            inputMode="url"
            autoCapitalize="none"
            spellCheck={false}
            enterKeyHint="go"
            aria-invalid={urlFieldInvalid}
            aria-describedby={urlFieldInvalid ? "url-error" : undefined}
          />
          <button
            className="btn btn--primary composer__submit"
            type="submit"
            disabled={isLoading}
            aria-busy={isLoading}
          >
            {isLoading ? (
              <>
                <span className="spinner" aria-hidden="true" />
                Shortening…
              </>
            ) : (
              "Shorten"
            )}
          </button>
        </div>

        <div className="composer__foot">
          <button
            className="composer__options-toggle"
            type="button"
            aria-expanded={showOptions}
            aria-controls="composer-options"
            onClick={() => setShowOptions((open) => !open)}
          >
            <PlusIcon className={showOptions ? "icon is-open" : "icon"} />
            <span>Alias and expiry</span>
            {hasOptionValues ? (
              <span className="composer__options-dot" aria-hidden="true" />
            ) : null}
            <span className="sr-only">
              {hasOptionValues ? ", options set" : ""}
            </span>
          </button>

          <p className="composer__status">
            <span
              className={
                isAuthenticated
                  ? "composer__status-dot is-active"
                  : "composer__status-dot"
              }
              aria-hidden="true"
            />
            <span>{isAuthenticated ? "Signed in" : "Anonymous link"}</span>
            <button className="composer__status-action" type="button" onClick={onManage}>
              {isAuthenticated ? "Manage links" : "Sign in"}
            </button>
          </p>
        </div>

        <div
          id="composer-options"
          className="composer__options"
          data-open={showOptions}
          inert={!showOptions}
          aria-hidden={!showOptions}
        >
          <div className="composer__options-inner">
            <div className="composer__options-grid">
              <label className="field">
                <span className="field__label">Custom alias</span>
                <input
                  ref={aliasRef}
                  className="field__control"
                  type="text"
                  name="custom_alias"
                  placeholder="my-link"
                  maxLength={16}
                  value={alias}
                  onChange={(event) => handleAliasChange(event.target.value)}
                  onBlur={() => setAliasTouched(true)}
                  autoComplete="off"
                  autoCapitalize="none"
                  spellCheck={false}
                  aria-invalid={showAliasError}
                  aria-describedby="alias-hint"
                />
                <span
                  id="alias-hint"
                  className="field__hint"
                  data-tone={showAliasError ? "danger" : undefined}
                  role={aliasApiError ? "alert" : undefined}
                >
                  {aliasApiError
                    ? (error ?? "That custom alias is not available.")
                    : showAliasError
                      ? "3–16 characters: letters, numbers, hyphens, or underscores."
                      : "3–16 letters, numbers, hyphens, or underscores."}
                </span>
              </label>

              <label className="field">
                <span className="field__label">Expires</span>
                <select
                  className="field__control"
                  name="expires_hours"
                  value={expiryHours}
                  onChange={(event) => handleExpiryChange(event.target.value)}
                >
                  {EXPIRY_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              {expiryHours === "custom" ? (
                <label className="field composer__custom">
                  <span className="field__label">Hours until expiry</span>
                  <input
                    ref={customExpiryRef}
                    className="field__control"
                    type="number"
                    name="custom_expiry_hours"
                    min={1}
                    max={CUSTOM_HOURS_MAX}
                    step={1}
                    inputMode="numeric"
                    placeholder="24"
                    value={customHours}
                    onChange={(event) => handleCustomHoursChange(event.target.value)}
                    required
                    aria-invalid={Boolean(customError)}
                    aria-describedby="custom-expiry-hint"
                  />
                  <span
                    id="custom-expiry-hint"
                    className="field__hint"
                    data-tone={customError ? "danger" : undefined}
                  >
                    {customError ?? "Up to 8,760 hours (one year)."}
                  </span>
                </label>
              ) : null}
            </div>
          </div>
        </div>
      </div>

      {visibleError && !aliasApiError ? (
        <p className="form-error composer__error" id="url-error" role="alert">
          <AlertIcon className="icon" />
          <span>{visibleError}</span>
        </p>
      ) : null}
    </form>
  );
}
