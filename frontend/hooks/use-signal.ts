"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";

interface UseSignalOptions {
  slug: string;
  onSuccess?: () => void;
}

export function useSignal({ slug, onSuccess }: UseSignalOptions) {
  const [loading, setLoading] = useState(false);

  const send = useCallback(
    async (action: string, body?: Record<string, unknown>) => {
      setLoading(true);
      try {
        const res = await fetch(`/api/admin/companies/${slug}/signal/${action}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body ?? {}),
        });

        if (!res.ok) {
          const data = await res.json();
          throw new Error(data.error || "Failed to send signal");
        }

        toast.success(`Signal sent: ${action}`);
        if (onSuccess) onSuccess();
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to send signal");
      } finally {
        setLoading(false);
      }
    },
    [slug, onSuccess]
  );

  return { send, loading };
}
