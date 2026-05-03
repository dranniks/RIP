import type { Claim, ClaimRow } from "../types/domain";

const now = new Date();

const formatDateTime = (date: Date): string =>
  date.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  });

const pendingRow = (serviceSlug: string, quantity = 1): ClaimRow => ({
  serviceSlug,
  quantity,
  compositionResult: "Будет рассчитано после формирования заявки",
  matchScore: "-"
});

export const createInitialDraftClaim = (): Claim => ({
  claimCode: "34",
  status: "черновик",
  createdAt: formatDateTime(now),
  formedAt: "-",
  completedAt: "-",
  completionFormulaResult: "",
  input: {
    iCu: "0.79",
    iZn: "0.19",
    iSn: "0.01",
    iPb: "0.02",
    mm: "Комментарий оператора отсутствует"
  },
  rows: [pendingRow("silver-byzantium")]
});

export const createPendingRow = (serviceSlug: string): ClaimRow => pendingRow(serviceSlug);
