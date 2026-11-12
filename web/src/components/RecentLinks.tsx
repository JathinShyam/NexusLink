import { useEffect, useRef, useState } from "react";
import { copyText } from "../lib/clipboard";
import {
  formatDateTime,
  formatRelativeTime,
  formatShortDate,
  isExpired,
} from "../lib/format";
import { displayUrl } from "../lib/url";
import type { RecentLink } from "../hooks/useRecentLinks";
import { AlertIcon, CheckIcon, CloseIcon, CopyIcon } from "./Icons";
import "./RecentLinks.css";

type Props = {
  items: RecentLink[];
  onRemove: (shortCode: string) => void;
};

export function RecentLinks({ items, onRemove }: Props) {
  const [copiedCode, setCopiedCode] = useState<string | null>(null);
  const [message, setMessage] = useState("");
  const [copyError, setCopyError] = useState(false);
  const [now, setNow] = useState(Date.now);
  const copyTimer = useRef<number | undefined>(undefined);

  useEffect(() => {
    const interval = window.setInterval(() => setNow(Date.now()), 60_000);
    return () => {
      window.clearInterval(interval);
      window.clearTimeout(copyTimer.current);
    };
  }, []);

  async function copy(item: RecentLink) {
    window.clearTimeout(copyTimer.current);
    setCopyError(false);

    try {
      await copyText(item.short_url);
      setCopiedCode(item.short_code);
      setMessage(`Copied ${item.short_url}`);
      copyTimer.current = window.setTimeout(() => setCopiedCode(null), 2000);
    } catch {
      setCopiedCode(null);
      setCopyError(true);
      setMessage(`Could not copy ${item.short_url}`);
    }
  }

  function remove(item: RecentLink) {
    onRemove(item.short_code);
    setMessage(`Removed ${item.short_url} from history.`);
  }

  return (
    <section className="recent" aria-labelledby="recent-title">
      <div className="recent__head">
        <h2 id="recent-title" className="recent__title">
          Recent links
        </h2>
        <p className="recent__hint">Kept in this browser</p>
      </div>

      <ul className="recent__list">
        {items.map((item) => {
          const expired = isExpired(item.expires_at);
          const copied = copiedCode === item.short_code;
          const shortLabel = displayUrl(item.short_url);
          const slash = shortLabel.indexOf("/");
          const shortHost = slash === -1 ? shortLabel : shortLabel.slice(0, slash + 1);
          const shortRest = slash === -1 ? "" : shortLabel.slice(slash + 1);

          return (
            <li className="recent__item" key={item.short_code} data-expired={expired || undefined}>
              <span className="recent__node" aria-hidden="true" />

              <div className="recent__body">
                <a
                  className="recent__short"
                  href={item.short_url}
                  target="_blank"
                  rel="noreferrer"
                  title={item.short_url}
                  translate="no"
                >
                  {shortHost}
                  {shortRest ? (
                    <>
                      <wbr />
                      <span className="recent__code">{shortRest}</span>
                    </>
                  ) : null}
                </a>
                <p className="recent__long" title={item.long_url}>
                  <span translate="no">{displayUrl(item.long_url)}</span>
                </p>
              </div>

              <div className="recent__meta">
                <time dateTime={item.created_at} title={formatDateTime(item.created_at)}>
                  {formatRelativeTime(item.created_at, now)}
                </time>
                {item.expires_at ? (
                  <span
                    className="recent__tag"
                    data-expired={expired || undefined}
                    title={`Expires ${formatDateTime(item.expires_at)}`}
                  >
                    {expired ? "Expired" : `Expires ${formatShortDate(item.expires_at)}`}
                  </span>
                ) : null}
              </div>

              <div className="recent__actions">
                <button
                  className="recent__action"
                  type="button"
                  onClick={() => void copy(item)}
                  aria-label={copied ? `Copied ${item.short_url}` : `Copy ${item.short_url}`}
                  title={copied ? "Copied" : "Copy link"}
                >
                  {copied ? <CheckIcon className="icon" /> : <CopyIcon className="icon" />}
                </button>
                <button
                  className="recent__action"
                  type="button"
                  onClick={() => remove(item)}
                  aria-label={`Remove ${item.short_url} from history`}
                  title="Remove from history"
                >
                  <CloseIcon className="icon" />
                </button>
              </div>
            </li>
          );
        })}
      </ul>

      <p className="sr-only" role="status" aria-live="polite">
        {message}
      </p>

      {copyError ? (
        <p className="form-error recent__error" role="alert">
          <AlertIcon className="icon" />
          <span>Copy failed. Select the link and copy it manually.</span>
        </p>
      ) : null}
    </section>
  );
}
