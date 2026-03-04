"use client";

import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

interface ElapsedTimerProps {
  startedAt: string;
  size?: "sm" | "lg";
  className?: string;
}

function computeElapsed(startedAt: string) {
  const start = new Date(startedAt).getTime();
  const now = Date.now();
  const diff = Math.max(0, now - start);

  const totalSeconds = Math.floor(diff / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  return { days, hours, minutes, seconds };
}

export function ElapsedTimer({ startedAt, size = "sm", className }: ElapsedTimerProps) {
  const [elapsed, setElapsed] = useState(() => computeElapsed(startedAt));

  useEffect(() => {
    setElapsed(computeElapsed(startedAt));
    const id = setInterval(() => setElapsed(computeElapsed(startedAt)), 1000);
    return () => clearInterval(id);
  }, [startedAt]);

  const pad = (n: number) => String(n).padStart(2, "0");

  if (size === "lg") {
    return (
      <div className={cn("flex items-baseline gap-3 font-mono tabular-nums", className)}>
        <div className="flex items-baseline gap-1.5">
          <span className="text-5xl font-bold tracking-tight">{elapsed.days}</span>
          <span className="text-lg text-muted-foreground">days</span>
        </div>
        <span className="text-3xl font-light text-muted-foreground/60">:</span>
        <span className="text-3xl font-medium tracking-tight">
          {pad(elapsed.hours)}:{pad(elapsed.minutes)}:{pad(elapsed.seconds)}
        </span>
      </div>
    );
  }

  return (
    <span className={cn("font-mono text-sm tabular-nums tracking-tight", className)}>
      {elapsed.days}d {pad(elapsed.hours)}:{pad(elapsed.minutes)}:{pad(elapsed.seconds)}
    </span>
  );
}
