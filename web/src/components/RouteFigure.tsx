import "./RouteFigure.css";

const LONG_URL = "https://example.com/pricing?utm_source=newsletter";

/* Six ways in, one way out. Each path is a different long address; they merge
   at the junction and leave as a single short code. */
const ROUTES = [
  "M8 74H132C196 74 214 159 360 159",
  "M8 108H186C240 108 252 159 360 159",
  "M8 142H108C176 142 188 159 360 159",
  "M8 176H214C256 176 262 159 360 159",
  "M8 210H152C212 210 226 159 360 159",
  "M8 244H198C244 244 258 159 360 159",
];

export function RouteFigure() {
  return (
    <figure className="route-figure">
      <svg
        className="route-figure__canvas"
        viewBox="0 0 520 300"
        role="img"
        aria-labelledby="route-figure-title route-figure-desc"
      >
        <title id="route-figure-title">Many long routes collapse into one code</title>
        <desc id="route-figure-desc">
          {LONG_URL} is shortened to the code aZ39kT.
        </desc>

        <text className="route-figure__url" x="8" y="34">
          {LONG_URL}
        </text>

        {ROUTES.map((d, index) => (
          <path
            className="route-figure__route"
            d={d}
            key={d}
            pathLength={1}
            style={{ animationDelay: `${index * 70}ms` }}
          />
        ))}

        <path className="route-figure__link" d="M360 159H386" pathLength={1} />

        <circle className="route-figure__halo" cx="360" cy="159" r="14" />
        <circle className="route-figure__node" cx="360" cy="159" r="6.5" />

        <rect className="route-figure__chip" x="386" y="142" width="118" height="34" rx="6" />
        <text className="route-figure__code" x="445" y="165" textAnchor="middle">
          aZ39kT
        </text>
      </svg>

      <figcaption className="route-figure__caption">
        One request in, one code out.
      </figcaption>
    </figure>
  );
}
