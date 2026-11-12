import type { ReactNode } from "react";
import "./AuthPages.css";

type Point = {
  term: string;
  copy: string;
};

type Props = {
  title: ReactNode;
  lede: string;
  points: Point[];
  children: ReactNode;
};

export function AuthLayout({ title, lede, points, children }: Props) {
  return (
    <main id="main" className="auth-main" tabIndex={-1}>
      <div className="auth-page shell">
        <section className="auth-intro">
          <h1 className="auth-title">{title}</h1>
          <p className="auth-lede">{lede}</p>

          <dl className="auth-points">
            {points.map((point) => (
              <div className="auth-point" key={point.term}>
                <dt>{point.term}</dt>
                <dd>{point.copy}</dd>
              </div>
            ))}
          </dl>
        </section>

        <section className="auth-panel">{children}</section>
      </div>
    </main>
  );
}
