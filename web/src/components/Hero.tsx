import "./Hero.css";

export function Hero() {
  return (
    <header className="hero">
      <div className="hero__plane" aria-hidden="true">
        <svg className="hero__paths" viewBox="0 0 1200 720" preserveAspectRatio="xMidYMid slice">
          <defs>
            <linearGradient id="pathGrad" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#2DD4BF" stopOpacity="0.55" />
              <stop offset="100%" stopColor="#2DD4BF" stopOpacity="0.05" />
            </linearGradient>
          </defs>
          <path
            d="M-40 420 C 180 280, 320 520, 520 360 S 820 180, 1040 300 S 1240 480, 1320 400"
            fill="none"
            stroke="url(#pathGrad)"
            strokeWidth="2.5"
          />
          <path
            d="M-20 560 C 220 480, 380 640, 560 500 S 860 360, 1100 480 S 1280 620, 1340 560"
            fill="none"
            stroke="#2DD4BF"
            strokeOpacity="0.18"
            strokeWidth="1.5"
          />
          <circle cx="520" cy="360" r="5" fill="#2DD4BF" fillOpacity="0.9" />
          <circle cx="820" cy="240" r="4" fill="#2DD4BF" fillOpacity="0.55" />
          <circle cx="1040" cy="300" r="6" fill="#E8EEF7" fillOpacity="0.35" />
        </svg>
        <div className="hero__grid" />
      </div>

      <div className="hero__content">
        <p className="hero__brand anim-fade-up">NexusLink</p>
        <h1 className="hero__headline anim-fade-up-delay">Short links that scale like infrastructure.</h1>
        <p className="hero__support anim-fade-up-delay">
          Paste a URL. Get a compact redirect. Own it with a token when you need to.
        </p>
        <div className="hero__cta anim-fade-up-delay">
          <a className="btn btn--primary" href="#shorten">
            Shorten a link
          </a>
          <a className="btn btn--ghost" href="#manage">
            Manage with JWT
          </a>
        </div>
      </div>
    </header>
  );
}
