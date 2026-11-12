import "./SpecStrip.css";

const FACTS = [
  "Base62 short codes",
  "Redis-cached redirects",
  "30 writes/min per IP",
  "JWT sessions",
  "Kafka to ClickHouse",
  "OpenAPI 3.0 contract",
];

export function SpecStrip() {
  return (
    <section className="spec-strip" aria-label="What NexusLink runs on">
      <ul className="spec-strip__list shell">
        {FACTS.map((fact) => (
          <li className="spec-strip__item" key={fact}>
            {fact}
          </li>
        ))}
      </ul>
    </section>
  );
}
