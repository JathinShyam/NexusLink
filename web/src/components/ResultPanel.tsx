import { useEffect, useRef, useState } from "react";
import type { ShortenResponse } from "../api/client";
import { copyText } from "../lib/clipboard";
import { formatDateTime } from "../lib/format";
import { CheckIcon, CopyIcon, ExternalIcon } from "./Icons";
import "./ResultPanel.css";

type Props = {
  result: ShortenResponse;
  onReset: () => void;
};

function splitShortUrl(shortUrl: string): { host: string; code: string } {
  try {
    const url = new URL(shortUrl);
    return {
      host: `${url.protocol}//${url.host}/`,
      code: url.pathname.replace(/^\//, ""),
    };
  } catch {
    return { host: "", code: shortUrl };
  }
}

export function ResultPanel({ result, onReset }: Props) {
  const [copied, setCopied] = useState(false);
  const [copyFailed, setCopyFailed] = useState(false);
  const copyButtonRef = useRef<HTMLButtonElement>(null);
  const timerRef = useRef<number | undefined>(undefined);
  const { host, code } = splitShortUrl(result.short_url);

  useEffect(() => {
    copyButtonRef.current?.focus();

    return () => {
      window.clearTimeout(timerRef.current);
    };
  }, []);

  async function copy() {
    window.clearTimeout(timerRef.current);
    setCopyFailed(false);

    try {
      await copyText(result.short_url);
      setCopied(true);
      timerRef.current = window.setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
      setCopyFailed(true);
      timerRef.current = window.setTimeout(() => setCopyFailed(false), 3000);
    }
  }

  return (
    <section className="result" aria-labelledby="result-title">
      <p className="sr-only" role="status" aria-live="polite">
        Short link created: {result.short_url}
      </p>

      <div className="result__head">
        <h2 id="result-title" className="result__label">
          Short link
        </h2>
        <p className="result__state">
          <CheckIcon className="icon" />
          <span>Link ready</span>
        </p>
      </div>

      <div className="result__row">
        <a
          className="result__url"
          href={result.short_url}
          target="_blank"
          rel="noreferrer"
          translate="no"
        >
          {host ? <span className="result__host">{host}</span> : null}
          <span className="result__code">{code}</span>
        </a>

        <button
          ref={copyButtonRef}
          className="btn btn--primary result__copy"
          type="button"
          onClick={copy}
        >
          {copied ? <CheckIcon className="icon" /> : <CopyIcon className="icon" />}
          {copied ? "Copied" : "Copy link"}
        </button>
      </div>

      {copyFailed ? (
        <p className="form-error result__copy-error" role="alert">
          Copy failed. Select the link above and copy it manually.
        </p>
      ) : null}

      <p className="sr-only" role="status" aria-live="polite">
        {copied
          ? "Short link copied to clipboard."
          : copyFailed
            ? "Could not copy the short link."
            : ""}
      </p>

      <div className="result__perf" aria-hidden="true" />

      <dl className="result__meta">
        <div>
          <dt>Points to</dt>
          <dd translate="no">{result.long_url}</dd>
        </div>
        <div>
          <dt>Expires</dt>
          <dd>
            {result.expires_at ? (
              <time dateTime={result.expires_at}>{formatDateTime(result.expires_at)}</time>
            ) : (
              "Never"
            )}
          </dd>
        </div>
      </dl>

      <div className="result__actions">
        <button
          className="btn btn--quiet result__reset"
          type="button"
          onClick={onReset}
        >
          Shorten another
        </button>
        <a
          className="btn btn--ghost result__open"
          href={result.short_url}
          target="_blank"
          rel="noreferrer"
        >
          Open link
          <ExternalIcon className="icon" />
        </a>
      </div>
    </section>
  );
}
