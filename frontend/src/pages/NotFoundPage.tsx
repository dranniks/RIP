import { Link } from "react-router-dom";

export const NotFoundPage = (): JSX.Element => (
  <section className="card">
    <h1 className="page-title">404</h1>
    <p className="page-subtitle">Страница не найдена.</p>
    <p className="small">
      <Link to="/services">Перейти к списку услуг</Link>
    </p>
  </section>
);
