import type { ApiStatus } from "../../types/domain";

interface ApiStatusLineProps {
  status: ApiStatus;
}

const statusMap: Record<ApiStatus, string> = {
  checking: "Проверка соединения с API...",
  online: "API: соединение активно",
  offline: "API: backend недоступен (работаем в mock-режиме)"
};

export const ApiStatusLine = ({ status }: ApiStatusLineProps): JSX.Element => (
  <p className={`small api-state api-state-${status}`}>{statusMap[status]}</p>
);
