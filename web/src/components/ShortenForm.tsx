import { useState, type FormEvent } from "react";
import type { ShortenResponse } from "../api/client";
import "./ShortenForm.css";

type Props = {
  isLoading: boolean;
  error: string | null;
  hasToken: boolean;
  onSubmit: (values: {
    long_url: string;
    custom_alias?: string;
    expires_in_seconds?: number;
  }) => Promise<ShortenResponse | null>;
};

export function ShortenForm({ isLoading, error, hasToken, onSubmit }: Props) {
  const [longUrl, setLongUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [expiryHours, setExpiryHours] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const hours = Number(expiryHours);
    await onSubmit({
      long_url: longUrl,
      custom_alias: alias || undefined,
      expires_in_seconds: hours > 0 ? Math.round(hours * 3600) : undefined,
    });
  }

  return (
    <section id="shorten" className="panel shorten" aria-labelledby="shorten-title">
      <div className="panel__head">
        <h2 id="shorten-title" className="panel__title">
          Create short link
        </h2>
        <p className="panel__lede">
          {hasToken
            ? "JWT attached — new links are owned and can be deleted later."
            : "Anonymous by default. Add a JWT below to own what you create."}
        </p>
      </div>

      <form className="shorten__form" onSubmit={handleSubmit}>
        <label className="field">
          <span className="field__label">Long URL</span>
          <input
            className="field__input"
            type="url"
            name="long_url"
            required
            placeholder="https://example.com/docs/path"
            value={longUrl}
            onChange={(e) => setLongUrl(e.target.value)}
            autoComplete="url"
          />
        </label>

        <div className="shorten__row">
          <label className="field">
            <span className="field__label">Custom alias</span>
            <input
              className="field__input"
              type="text"
              name="custom_alias"
              placeholder="my-link"
              maxLength={16}
              pattern="[A-Za-z0-9_-]{3,16}"
              title="3–16 alphanumeric characters, hyphens, or underscores"
              value={alias}
              onChange={(e) => setAlias(e.target.value)}
            />
          </label>

          <label className="field">
            <span className="field__label">Expires in (hours)</span>
            <input
              className="field__input"
              type="number"
              name="expires_hours"
              min={1}
              step={1}
              placeholder="optional"
              value={expiryHours}
              onChange={(e) => setExpiryHours(e.target.value)}
            />
          </label>
        </div>

        {error ? (
          <p className="form-error" role="alert">
            {error}
          </p>
        ) : null}

        <button className="btn btn--primary shorten__submit" type="submit" disabled={isLoading}>
          {isLoading ? "Creating…" : "Shorten"}
        </button>
      </form>
    </section>
  );
}
