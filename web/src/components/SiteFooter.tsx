import "./SiteFooter.css";

export function SiteFooter() {
  return (
    <footer className="footer">
      <div className="footer__inner">
        <p className="footer__brand">NexusLink</p>
        <p className="footer__note">
          Local demo UI for the Go API — Redis cache, rate limits, JWT delete, Prometheus metrics.
        </p>
        <p className="footer__meta">
          API <code>localhost:8080</code> · UI <code>localhost:5173</code>
        </p>
      </div>
    </footer>
  );
}
