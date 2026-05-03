import { Link } from "react-router-dom";

export const SiteHeader = (): JSX.Element => (
  <header className="header">
    <Link className="home-link" to="/services" aria-label="Домой">
      <svg viewBox="0 0 24 24">
        <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" />
      </svg>
    </Link>
    <p className="site-name">ArchaeoXRF Atlas: XRF анализ по частотам и интенсивности.</p>
  </header>
);
