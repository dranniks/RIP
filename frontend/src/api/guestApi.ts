import { apiRequest } from "./client";

interface HealthResponse {
  status: string;
}

export const pingBackend = async (): Promise<boolean> => {
  try {
    await apiRequest<HealthResponse>("/health");
    return true;
  } catch {
    return false;
  }
};

export const logPageVisit = async (page: string): Promise<void> => {
  try {
    await apiRequest<void>("/frontend/visit", {
      method: "POST",
      body: JSON.stringify({
        page,
        visitedAt: new Date().toISOString()
      })
    });
  } catch {
    // В mock-режиме backend может не иметь этот endpoint.
  }
};
