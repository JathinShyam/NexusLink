import type { ApiStatus } from "../hooks/useApiStatus";
import "./SiteFooter.css";

type Props = {
  status: ApiStatus;
};

const STATUS_LABEL: Record<ApiStatus, string> = {
  checking: "Checking API…",
  ready: "API ready",
  unavailable: "API unavailable",
};

export function SiteFooter({ status }: Props) {
  return (
    <footer className="site-footer">
      <div className="site-footer__inner shell">
        <p className={`site-footer__status is-${status}`} role="status" aria-live="polite">
          <span className="site-footer__dot" aria-hidden="true" />
          {STATUS_LABEL[status]}
        </p>
        <p className="site-footer__note">Your session stays in this browser.</p>
      </div>
    </footer>
  );
}
