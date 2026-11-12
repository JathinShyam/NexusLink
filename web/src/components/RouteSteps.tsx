import { useEffect, useRef, useState } from "react";
import "./RouteSteps.css";

const STEPS = [
  {
    index: "01",
    title: "Paste",
    copy: "Any http or https address. It is normalized on the way in, and private hosts are refused.",
  },
  {
    index: "02",
    title: "Route",
    copy: "The URL becomes a base62 code, cached in Redis, and answered with a single redirect.",
  },
  {
    index: "03",
    title: "Share",
    copy: "Copy it, embed it, or call the API. Expiry and ownership are set before it leaves.",
  },
];

export function RouteSteps() {
  const sectionRef = useRef<HTMLElement>(null);
  const [drawn, setDrawn] = useState(false);

  useEffect(() => {
    const node = sectionRef.current;
    if (!node) {
      return;
    }

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setDrawn(true);
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setDrawn(true);
            observer.disconnect();
          }
        }
      },
      { threshold: 0.4 },
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  return (
    <section className="section route" ref={sectionRef} data-drawn={drawn}>
      <div className="section__head shell">
        <h2 className="section__title">Three steps, one request.</h2>
        <p className="section__lede">
          A short link is the shortest path between a page you cannot post and a
          link you can. Everything else in NexusLink exists to keep that path
          honest.
        </p>
      </div>

      <ol className="route__steps shell">
        {STEPS.map((step) => (
          <li className="route__step" key={step.index}>
            <span className="route__node" aria-hidden="true" />
            <span className="route__index">{step.index}</span>
            <h3 className="route__title">{step.title}</h3>
            <p className="route__copy">{step.copy}</p>
          </li>
        ))}
      </ol>
    </section>
  );
}
