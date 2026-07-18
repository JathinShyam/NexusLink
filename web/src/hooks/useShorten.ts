import { useCallback, useState } from "react";
import {
  ApiError,
  friendlyError,
  shorten,
  type ShortenRequest,
  type ShortenResponse,
} from "../api/client";

type Status = "idle" | "loading" | "success" | "error";

export function useShorten(token: string) {
  const [status, setStatus] = useState<Status>("idle");
  const [result, setResult] = useState<ShortenResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const create = useCallback(
    async (input: ShortenRequest) => {
      setStatus("loading");
      setError(null);
      try {
        const payload: ShortenRequest = {
          long_url: input.long_url.trim(),
        };
        const alias = input.custom_alias?.trim();
        if (alias) {
          payload.custom_alias = alias;
        }
        if (input.expires_in_seconds && input.expires_in_seconds > 0) {
          payload.expires_in_seconds = input.expires_in_seconds;
        }

        const response = await shorten(payload, token || null);
        setResult(response);
        setStatus("success");
        return response;
      } catch (err) {
        setResult(null);
        setError(friendlyError(err));
        setStatus("error");
        if (!(err instanceof ApiError) && !(err instanceof Error)) {
          console.error(err);
        }
        return null;
      }
    },
    [token],
  );

  const reset = useCallback(() => {
    setStatus("idle");
    setResult(null);
    setError(null);
  }, []);

  return { status, result, error, create, reset, isLoading: status === "loading" };
}
