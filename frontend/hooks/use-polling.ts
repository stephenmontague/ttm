"use client";

import { useState, useEffect, useCallback, useRef } from "react";

interface UsePollingOptions<T> {
  fetcher: () => Promise<T>;
  /** If set, poll continuously at this interval (ms). If omitted, fetch once on mount only. */
  interval?: number;
}

interface UsePollingResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  /** Poll every `intervalMs` until data changes or `maxMs` elapses. */
  pollUntilChanged: (intervalMs?: number, maxMs?: number) => void;
}

export function usePolling<T>({
  fetcher,
  interval,
}: UsePollingOptions<T>): UsePollingResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;
  const dataRef = useRef<T | null>(null);
  const burstTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const refresh = useCallback(async () => {
    try {
      const result = await fetcherRef.current();
      setData(result);
      dataRef.current = result;
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to fetch data");
    } finally {
      setLoading(false);
    }
  }, []);

  const pollUntilChanged = useCallback(
    (intervalMs = 1000, maxMs = 5000) => {
      // Clear any existing burst poll.
      if (burstTimerRef.current) clearInterval(burstTimerRef.current);

      const snapshot = JSON.stringify(dataRef.current);
      const deadline = Date.now() + maxMs;

      burstTimerRef.current = setInterval(async () => {
        if (Date.now() > deadline) {
          clearInterval(burstTimerRef.current!);
          burstTimerRef.current = null;
          return;
        }
        try {
          const result = await fetcherRef.current();
          setData(result);
          dataRef.current = result;
          setError(null);
          if (JSON.stringify(result) !== snapshot) {
            clearInterval(burstTimerRef.current!);
            burstTimerRef.current = null;
          }
        } catch {
          // Keep trying until deadline.
        }
      }, intervalMs);
    },
    []
  );

  // Fetch on mount, and optionally poll continuously.
  useEffect(() => {
    refresh();
    const id = interval ? setInterval(refresh, interval) : null;
    return () => {
      if (id) clearInterval(id);
      if (burstTimerRef.current) clearInterval(burstTimerRef.current);
    };
  }, [refresh, interval]);

  return { data, loading, error, refresh, pollUntilChanged };
}
