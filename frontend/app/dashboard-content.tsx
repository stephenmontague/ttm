"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { WorkflowCard } from "@/components/workflow-card";
import { usePolling } from "@/hooks/use-polling";
import type { Company } from "@/lib/types";
import { Activity, Trophy } from "lucide-react";

interface DashboardContentProps {
  initialCompanies: Company[];
}

async function fetchCompanies(): Promise<Company[]> {
  const res = await fetch("/api/companies");
  if (!res.ok) return [];
  const data = await res.json();
  return data.companies || [];
}

export function DashboardContent({ initialCompanies }: DashboardContentProps) {
  const { data: companies, loading } = usePolling<Company[]>({
    fetcher: fetchCompanies,
    interval: 30000,
  });

  const list = companies ?? initialCompanies;
  const active = list.filter((c) => c.Status === "active");
  const won = list.filter((c) => c.Status === "meeting_booked");

  return (
    <>
      {/* Hero */}
      <section className="relative border-b overflow-hidden">
        <div className="hero-gradient absolute inset-0" />
        <div className="hero-grid absolute inset-0" />
        <div className="relative mx-auto max-w-6xl px-6 py-20">
          <div className="max-w-2xl">
            <p className="mb-3 text-sm font-medium tracking-wider uppercase text-primary">
              Durable Execution in Action
            </p>
            <h1 className="text-5xl font-bold tracking-tight">
              <span className="text-gradient">Time to Meeting</span>
            </h1>
            <p className="mt-4 text-lg leading-relaxed text-muted-foreground">
              Tracking outreach cadence to target companies with{" "}
              <span className="font-semibold text-foreground">
                Temporal durable workflows
              </span>
              . Each company gets a long-running workflow that persists
              indefinitely — surviving restarts, deployments, and failures —
              until a meeting is booked.
            </p>
          </div>
        </div>
      </section>

      {/* Active Workflows */}
      <section className="mx-auto max-w-6xl px-6 py-12">
        <div className="flex items-center gap-2 mb-6">
          <Activity className="h-5 w-5 text-emerald-600 dark:text-emerald-400" />
          <h2 className="text-xl font-semibold tracking-tight">
            Active Workflows
          </h2>
          {active.length > 0 && (
            <span className="ml-2 rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-medium text-emerald-700 dark:text-emerald-400">
              {active.length}
            </span>
          )}
        </div>

        {loading && list.length === 0 ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-48 rounded-xl" />
            ))}
          </div>
        ) : active.length === 0 ? (
          <div className="rounded-xl border border-dashed p-12 text-center">
            <p className="text-sm text-muted-foreground">
              No active workflows right now. Check back soon.
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {active.map((company, index) => (
              <WorkflowCard key={company.ID} company={company} index={index} />
            ))}
          </div>
        )}
      </section>

      {/* Wall of Wins */}
      {won.length > 0 && (
        <section className="mx-auto max-w-6xl px-6 pb-12">
          <div className="flex items-center gap-2 mb-6">
            <Trophy className="h-5 w-5 text-primary" />
            <h2 className="text-xl font-semibold tracking-tight">
              Wall of Wins
            </h2>
            <span className="ml-2 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">
              {won.length}
            </span>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {won.map((company, index) => (
              <WorkflowCard key={company.ID} company={company} index={index} />
            ))}
          </div>
        </section>
      )}
    </>
  );
}
