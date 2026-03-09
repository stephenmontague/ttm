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
      <div className={cn("timer-glow rounded-xl border bg-card/80 p-6", className)}>
        <div className="flex items-end justify-center gap-1 font-mono tabular-nums">
          <div className="flex flex-col items-center">
            <span className="text-6xl font-bold tracking-tight text-gradient">{elapsed.days}</span>
            <span className="mt-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">Days</span>
          </div>
          <span className="timer-colon text-3xl font-light text-primary/40 mb-3">:</span>
          <div className="flex flex-col items-center">
            <span className="text-4xl font-semibold tracking-tight text-primary">{pad(elapsed.hours)}</span>
            <span className="mt-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">Hrs</span>
          </div>
          <span className="timer-colon text-3xl font-light text-primary/40 mb-3">:</span>
          <div className="flex flex-col items-center">
            <span className="text-4xl font-semibold tracking-tight text-primary">{pad(elapsed.minutes)}</span>
            <span className="mt-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">Min</span>
          </div>
          <span className="timer-colon text-3xl font-light text-primary/40 mb-3">:</span>
          <div className="flex flex-col items-center">
            <span className="text-4xl font-semibold tracking-tight text-primary">{pad(elapsed.seconds)}</span>
            <span className="mt-1 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">Sec</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <span className={cn("font-mono text-sm tabular-nums tracking-tight text-primary", className)}>
      {elapsed.days}d {pad(elapsed.hours)}:{pad(elapsed.minutes)}:{pad(elapsed.seconds)}
    </span>
  );
}
