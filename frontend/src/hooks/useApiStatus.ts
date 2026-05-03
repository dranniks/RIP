import { useEffect, useState } from "react";

import { logPageVisit, pingBackend } from "../api/guestApi";
import type { ApiStatus } from "../types/domain";

export const useApiStatus = (pageName: string): ApiStatus => {
  const [status, setStatus] = useState<ApiStatus>("checking");

  useEffect(() => {
    let isMounted = true;

    const syncWithApi = async (): Promise<void> => {
      setStatus("checking");
      await logPageVisit(pageName);
      const isOnline = await pingBackend();

      if (isMounted) {
        setStatus(isOnline ? "online" : "offline");
      }
    };

    void syncWithApi();

    return () => {
      isMounted = false;
    };
  }, [pageName]);

  return status;
};
