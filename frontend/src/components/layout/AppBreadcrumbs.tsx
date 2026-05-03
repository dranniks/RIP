import { Link, useLocation } from "react-router-dom";

import type { Service } from "../../types/domain";

interface BreadcrumbItem {
  label: string;
  to: string;
}

interface AppBreadcrumbsProps {
  services: Service[];
}

const baseItem: BreadcrumbItem = {
  label: "Список услуг",
  to: "/services"
};

const createBreadcrumbs = (pathname: string, services: Service[]): BreadcrumbItem[] => {
  if (pathname === "/" || pathname === "/services") {
    return [baseItem];
  }

  if (pathname.startsWith("/services/")) {
    const serviceIdentifier = pathname.replace("/services/", "");
    const maybeID = Number(serviceIdentifier);
    const service = Number.isFinite(maybeID)
      ? services.find((item) => item.id === maybeID)
      : services.find((item) => item.slug === serviceIdentifier);

    return [
      baseItem,
      {
        label: service?.name ?? `Услуга ${serviceIdentifier}`,
        to: pathname
      }
    ];
  }

  return [
    baseItem,
    {
      label: "Страница не найдена",
      to: pathname
    }
  ];
};

export const AppBreadcrumbs = ({ services }: AppBreadcrumbsProps): JSX.Element => {
  const location = useLocation();
  const crumbs = createBreadcrumbs(location.pathname, services);

  return (
    <nav className="breadcrumbs" aria-label="Путь до текущей страницы">
      {crumbs.map((crumb, index) => {
        const isLast = index === crumbs.length - 1;

        return (
          <span className="breadcrumbs-item" key={crumb.to}>
            {isLast ? <span>{crumb.label}</span> : <Link to={crumb.to}>{crumb.label}</Link>}
            {!isLast && <span className="breadcrumbs-separator">/</span>}
          </span>
        );
      })}
    </nav>
  );
};
