import { useEffect, useRef, useState } from "react";
import { copyText } from "../lib/clipboard";
import { CheckIcon, CopyIcon } from "./Icons";
import "./ApiPanel.css";

const ENDPOINTS = [
  "POST   /api/v1/shorten",
  "GET    /{short_code}",
  "DELETE /api/v1/links/{short_code}",
  "POST   /api/v1/auth/login",
];

export function ApiPanel() {
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    return () => {
      window.clearTimeout(timerRef.current);
    };
  }, []);

  async function copy() {
    window.clearTimeout(timerRef.current);
    try {
      await copyText(
        `curl -X POST http://localhost:8080/api/v1/shorten \\\n  -H 'Content-Type: application/json' \\\n  -d '{"long_url":"https://example.com/pricing"}'`,
      );
      setCopied(true);
      timerRef.current = window.setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  }

  return (
    <section className="section api" aria-labelledby="api-title">
      <div className="api__inner shell">
        <div className="api__intro">
          <h2 className="section__title" id="api-title">
            Ship it from your stack.
          </h2>
          <p className="section__lede">
            One request returns a short link. There is no SDK to install and no
            client library to keep in sync — the contract is the API.
          </p>

          <ul className="api__endpoints">
            {ENDPOINTS.map((endpoint) => (
              <li className="api__endpoint" key={endpoint}>
                {endpoint}
              </li>
            ))}
          </ul>
        </div>

        <div className="code">
          <div className="code__bar">
            <span className="code__label">Request</span>
            <button
              className="code__copy"
              type="button"
              onClick={() => void copy()}
              data-copied={copied ? "true" : undefined}
            >
              {copied ? <CheckIcon className="icon" /> : <CopyIcon className="icon" />}
              {copied ? "Copied" : "Copy"}
            </button>
          </div>

          <pre className="code__block">
            <code>
              <span className="code__method">curl</span>
              {" -X "}
              <span className="code__method">POST</span>
              {" http://localhost:8080/api/v1/shorten\n"}
              {"  "}
              {"-H "}
              <span className="code__string">&apos;Content-Type: application/json&apos;</span>
              {"\n  "}
              {"-d "}
              <span className="code__string">
                &apos;{"{"}long_url:&quot;https://example.com/pricing&quot;{"}"}&apos;
              </span>
            </code>
          </pre>

          <div className="code__bar code__bar--second">
            <span className="code__label">Response</span>
            <span className="code__status">201</span>
          </div>

          <pre className="code__block code__block--response">
            <code>
              {"{\n  "}
              <span className="code__key">&quot;short_code&quot;</span>
              {": "}
              <span className="code__value">&quot;aZ39kT&quot;</span>
              {",\n  "}
              <span className="code__key">&quot;short_url&quot;</span>
              {": "}
              <span className="code__value">&quot;http://localhost:8080/aZ39kT&quot;</span>
              {",\n  "}
              <span className="code__key">&quot;long_url&quot;</span>
              {": "}
              <span className="code__value">&quot;https://example.com/pricing&quot;</span>
              {"\n}"}
            </code>
          </pre>
        </div>
      </div>
    </section>
  );
}
