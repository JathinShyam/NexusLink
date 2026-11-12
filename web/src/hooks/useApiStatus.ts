import { useEffect, useState } from "react";
import { checkReadiness } from "../api/client";

export type ApiStatus = "checking" | "ready" | "unavailable";

export function useApiStatus(): ApiStatus {
  const [status, setStatus] = useState<ApiStatus>("checking");

  useEffect(() => {
    let active = true;
    let controller = new AbortController();

    async function check() {
      controller.abort();
      controller = new AbortController();
      const ready = await checkReadiness(controller.signal);

      if (active) {
        setStatus(ready ? "ready" : "unavailable");
      }
    }

    void check();

    function handleVisibility() {
      if (document.visibilityState === "visible") {
        void check();
      }
    }

    const interval = window.setInterval(check, 60_000);
    document.addEventListener("visibilitychange", handleVisibility);

    return () => {
      active = false;
      controller.abort();
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", handleVisibility);
    };
  }, []);

  return status;
}
