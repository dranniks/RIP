export interface Service {
  id: number;
  slug: string;
  name: string;
  era: string;
  culture: string;
  description: string;
  clipDescriptionEn: string;
  imageUrl: string;
  videoUrl?: string;
  cuReference: string;
  znReference: string;
  snReference: string;
  pbReference: string;
  price: number;
  availableDate: string;
}

export interface ClaimInputValues {
  iCu: string;
  iZn: string;
  iSn: string;
  iPb: string;
  mm: string;
}

export interface ClaimRow {
  serviceSlug: string;
  quantity: number;
  compositionResult: string;
  matchScore: string;
}

export type ClaimStatus = "черновик" | "сформирован" | "завершен" | "отклонен" | "удален";

export interface Claim {
  claimCode: string;
  status: ClaimStatus;
  createdAt: string;
  formedAt: string;
  completedAt: string;
  completionFormulaResult: string;
  input: ClaimInputValues;
  rows: ClaimRow[];
}

export interface ClaimSubmitPayload {
  claimCode: string;
  input: ClaimInputValues;
}

export type ApiStatus = "checking" | "online" | "offline";

export interface CartIcon {
  claimId: number | null;
  claimCode: string | null;
  serviceCount: number;
}
