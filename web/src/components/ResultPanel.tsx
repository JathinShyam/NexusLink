import { useState } from "react";
import type { ShortenResponse } from "../api/client";
import "./ResultPanel.css";

type Props = {
  result: ShortenResponse;
};

export function ResultPanel({ result }: Props) {
  const [copied, setCopied] = useState(false);
  const [flash, setFlash] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(result.short_url);
      setCopied(true);
      setFlash(true);
      window.setTimeout(() => setFlash(false), 700);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  }

  return (
    <section
      className={`panel result anim-slide-in ${flash ? "copy-flash" : ""}`}
      aria-live="polite"
      aria-labelledby="result-title"
    >
      <div className="panel__head">
        <h2 id="result-title" className="panel__title">
          Your short link
        </h2>
        <p className="panel__lede">Redirects with HTTP 302. Cache-warmed on create.</p>
      </div>

      <div className="result__url">
        <a href={result.short_url} target="_blank" rel="noreferrer">
          {result.short_url}
        </a>
      </div>

      <dl className="result__meta">
        <div>
          <dt>Code</dt>
          <dd>
            <code>{result.short_code}</code>
          </dd>
        </div>
        <div>
          <dt>Target</dt>
          <dd className="result__target">{result.long_url}</dd>
        </div>
        {result.expires_at ? (
          <div>
            <dt>Expires</dt>
            <dd>{new Date(result.expires_at).toLocaleString()}</dd>
          </div>
        ) : null}
      </dl>

      <div className="result__actions">
        <button className="btn btn--primary" type="button" onClick={copy}>
          {copied ? "Copied" : "Copy link"}
        </button>
        <a className="btn btn--ghost" href={result.short_url} target="_blank" rel="noreferrer">
          Open
        </a>
      </div>
    </section>
  );
}
