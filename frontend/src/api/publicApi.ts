import { mockServices, servicesByID } from "../mock/services";
import type { CartIcon, Service } from "../types/domain";
import { apiRequest } from "./client";

interface ApiEnvelope<T> {
  data: T;
}

interface BackendService {
  id: number;
  slug: string;
  name: string;
  description: string;
  clip_description_en?: string;
  image_url: string | null;
  video_url: string | null;
  era: string;
  culture: string;
  unit_price: number;
  cu_reference: number;
  zn_reference: number;
  sn_reference: number;
  pb_reference: number;
  created_at: string;
}

interface BackendCartIcon {
  claim_id: number | null;
  claim_code: string | null;
  service_count: number;
}

export interface ServiceFiltersRequest {
  query: string;
}

export interface ApiDataResponse<T> {
  data: T;
  source: "backend" | "mock";
  note?: string;
}

const asString = (value: number | string | null | undefined): string => {
  if (value === null || value === undefined) {
    return "";
  }
  return typeof value === "number" ? value.toString() : value;
};

const clipDescriptionsBySlug: Record<string, string> = {
  "alloy-bronze-cyprus":
    "Dark brown bronze ingot with oxidized rough texture and warm copper highlights under light.",
  "bronze-cyprus":
    "Dark brown bronze ingot with oxidized rough texture and warm copper highlights under light.",
  "alloy-brass-rome":
    "Warm yellow brass ingot with bright golden edges and medium metallic reflectance on surface.",
  "brass-rome":
    "Warm yellow brass ingot with bright golden edges and medium metallic reflectance on surface.",
  "alloy-iron-north":
    "Dark gray iron ingot with matte grainy texture, cold tone, and weak silver highlights.",
  "iron-north":
    "Dark gray iron ingot with matte grainy texture, cold tone, and weak silver highlights.",
  "alloy-silver-byzantium":
    "Pale silver ingot with smooth texture, strong reflectance, and bright white specular highlights.",
  "silver-byzantium":
    "Pale silver ingot with smooth texture, strong reflectance, and bright white specular highlights."
};

const buildFallbackClipDescription = (service: BackendService): string => {
  const normalizedSlug = service.slug.trim().toLowerCase();
  const bySlug = clipDescriptionsBySlug[normalizedSlug];

  if (bySlug) {
    return bySlug;
  }

  if (normalizedSlug.includes("silver")) {
    return "Ancient silver alloy artifact with bright gray metal surface for XRF similarity search.";
  }
  if (normalizedSlug.includes("brass")) {
    return "Ancient brass alloy artifact with yellow metallic surface for archaeological XRF matching.";
  }
  if (normalizedSlug.includes("bronze")) {
    return "Ancient bronze alloy artifact with brown metallic surface for archaeological XRF matching.";
  }
  if (normalizedSlug.includes("iron")) {
    return "Ancient iron alloy artifact with dark matte metal texture for archaeological XRF matching.";
  }

  return "Archaeological metal reference alloy sample for XRF spectral comparison and CLIP similarity search.";
};

const ensureClipLength = (text: string): string => {
  const normalized = text.replace(/\s+/g, " ").trim();
  if (normalized.length >= 50 && normalized.length <= 100) {
    return normalized;
  }

  if (normalized.length < 50) {
    return `${normalized} Prepared for archaeological alloy search in CLIP.`;
  }

  return normalized.slice(0, 100).trimEnd();
};

const toService = (raw: BackendService): Service => {
  const clipText =
    typeof raw.clip_description_en === "string" && raw.clip_description_en.trim().length > 0
      ? raw.clip_description_en
      : buildFallbackClipDescription(raw);

  return {
    id: raw.id,
    slug: raw.slug,
    name: raw.name,
    description: raw.description,
    clipDescriptionEn: ensureClipLength(clipText),
    imageUrl: raw.image_url ?? "",
    videoUrl: raw.video_url ?? undefined,
    era: raw.era,
    culture: raw.culture,
    price: raw.unit_price,
    availableDate: raw.created_at ?? "",
    cuReference: asString(raw.cu_reference),
    znReference: asString(raw.zn_reference),
    snReference: asString(raw.sn_reference),
    pbReference: asString(raw.pb_reference)
  };
};

const applyClientFilters = (services: Service[], filters: ServiceFiltersRequest): Service[] => {
  const query = filters.query.trim().toLowerCase();
  if (query.length === 0) {
    return services;
  }
  return services.filter((service) => service.name.toLowerCase().includes(query));
};

const buildServicesQuery = (filters: ServiceFiltersRequest): string => {
  const params = new URLSearchParams();
  if (filters.query.trim().length > 0) {
    params.set("q", filters.query.trim());
  }
  const query = params.toString();
  return query.length > 0 ? `?${query}` : "";
};

const getJwtToken = (): string | null => {
  const keys = ["jwt", "token", "auth_token", "jwt_token"];
  for (const key of keys) {
    const value = localStorage.getItem(key);
    if (value && value.trim().length > 0) {
      return value.trim();
    }
  }
  return null;
};

export const getServices = async (filters: ServiceFiltersRequest): Promise<ApiDataResponse<Service[]>> => {
  const query = buildServicesQuery(filters);

  try {
    const response = await apiRequest<ApiEnvelope<BackendService[]>>(`/services${query}`);
    const mapped = response.data.map(toService);
    return {
      data: applyClientFilters(mapped, filters),
      source: "backend"
    };
  } catch {
    return {
      data: applyClientFilters(mockServices, filters),
      source: "mock",
      note: "Бэкенд недоступен, показаны mock-данные."
    };
  }
};

export const getServiceByID = async (serviceID: number): Promise<ApiDataResponse<Service | null>> => {
  if (!Number.isFinite(serviceID) || serviceID <= 0) {
    return {
      data: null,
      source: "mock",
      note: "Некорректный ID услуги."
    };
  }

  try {
    const response = await apiRequest<ApiEnvelope<BackendService>>(`/services/${serviceID}`);
    return {
      data: toService(response.data),
      source: "backend"
    };
  } catch {
    return {
      data: servicesByID.get(serviceID) ?? null,
      source: "mock",
      note: "Карточка услуги взята из mock-данных."
    };
  }
};

export const getCartIcon = async (): Promise<ApiDataResponse<CartIcon>> => {
  const token = getJwtToken();
  const headers: HeadersInit = token ? { Authorization: `Bearer ${token}` } : {};

  try {
    const response = await apiRequest<ApiEnvelope<BackendCartIcon>>("/claims/cart-icon", {
      method: "GET",
      headers
    });

    return {
      data: {
        claimId: response.data.claim_id ?? null,
        claimCode: response.data.claim_code ?? null,
        serviceCount: response.data.service_count ?? 0
      },
      source: "backend"
    };
  } catch {
    return {
      data: {
        claimId: null,
        claimCode: null,
        serviceCount: 0
      },
      source: "mock",
      note: "Иконка корзины показана в mock-режиме."
    };
  }
};
