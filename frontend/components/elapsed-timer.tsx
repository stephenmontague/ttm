"use client";

import { useEffect, useState } from "react";

interface ElapsedTimerProps {
  startedAt: string;
  className?: string;
}

export function ElapsedTimer({ startedAt, className }: ElapsedTimerProps) {
  const [elapsed, setElapsed] = useState(() => getElapsed(startedAt));

  useEffect(() => {
    const interval = setInterval(() => {
      setElapsed(getElapsed(startedAt));
    }, 1000);
    return () => clearInterval(interval);
  }, [startedAt]);

  return (
    <div className={className}>
      <span className="tabular-nums text-2xl font-bold">{elapsed.days}d</span>
      <span className="tabular-nums text-lg text-muted-foreground ml-1">
        {elapsed.hours}h {elapsed.minutes}m {elapsed.seconds}s
      </span>
    </div>
  );
}

function getElapsed(startedAt: string) {
  const diff = Date.now() - new Date(startedAt).getTime();
  const seconds = Math.floor(diff / 1000) % 60;
  const minutes = Math.floor(diff / (1000 * 60)) % 60;
  const hours = Math.floor(diff / (1000 * 60 * 60)) % 24;
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  return { days, hours, minutes, seconds };
}
