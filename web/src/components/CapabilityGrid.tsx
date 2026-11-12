import "./CapabilityGrid.css";

const CAPABILITIES = [
  {
    label: "safety",
    title: "Private hosts refused",
    copy: "Only public http and https targets resolve. Loopback, private and link-local addresses are rejected before a code is ever issued.",
  },
  {
    label: "alias",
    title: "Readable aliases",
    copy: "Claim a code instead of taking a random one: 3 to 16 letters, numbers, hyphens or underscores.",
  },
  {
    label: "revoke",
    title: "Revocation",
    copy: "Delete a link you own and the very next request gets nothing back. The cache entry goes with it.",
  },
  {
    label: "api",
    title: "A documented contract",
    copy: "Every endpoint is described in an OpenAPI 3.0 spec, so clients, tests and gateways can be generated from it.",
  },
  {
    label: "analytics",
    title: "Click analytics",
    copy: "Each redirect publishes a click event to Kafka, and a consumer batches it into ClickHouse for reporting.",
  },
  {
    label: "limits",
    title: "Limits you can see",
    copy: "Shortening is capped per IP and the redirect path runs on a token bucket, so one client cannot flood the rest.",
  },
];

export function CapabilityGrid() {
  return (
    <section className="section capabilities" aria-labelledby="capabilities-title">
      <div className="section__head shell">
        <h2 className="section__title" id="capabilities-title">
          Built to be depended on.
        </h2>
        <p className="section__lede">
          A short link is infrastructure: it has to survive being pasted into a
          ticket, a campaign, and a build pipeline. These are the guarantees
          behind it.
        </p>
      </div>

      <ul className="capabilities__grid shell">
        {CAPABILITIES.map((item) => (
          <li className="capabilities__item" key={item.label}>
            <span className="capabilities__label">{item.label}</span>
            <h3 className="capabilities__title">{item.title}</h3>
            <p className="capabilities__copy">{item.copy}</p>
          </li>
        ))}
      </ul>
    </section>
  );
}
