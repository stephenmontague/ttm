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
      <section className="border-b bg-linear-to-b from-muted/30 to-background">
        <div className="mx-auto max-w-6xl px-6 py-16">
          <div className="max-w-2xl">
            <h1 className="text-4xl font-bold tracking-tight">
              Time to Meeting
            </h1>
            <p className="mt-3 text-lg text-muted-foreground">
              Tracking outreach cadence to target companies with{" "}
              <span className="font-medium text-foreground">
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
              No active workflows yet. Start one from the{" "}
              <a
                href="/admin"
                className="font-medium underline underline-offset-4 hover:text-foreground"
              >
                admin panel
              </a>
              .
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {active.map((company) => (
              <WorkflowCard key={company.ID} company={company} />
            ))}
          </div>
        )}
      </section>

      {/* Wall of Wins */}
      {won.length > 0 && (
        <section className="mx-auto max-w-6xl px-6 pb-12">
          <div className="flex items-center gap-2 mb-6">
            <Trophy className="h-5 w-5 text-blue-600 dark:text-blue-400" />
            <h2 className="text-xl font-semibold tracking-tight">
              Wall of Wins
            </h2>
            <span className="ml-2 rounded-full bg-blue-500/10 px-2.5 py-0.5 text-xs font-medium text-blue-700 dark:text-blue-400">
              {won.length}
            </span>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {won.map((company) => (
              <WorkflowCard key={company.ID} company={company} />
            ))}
          </div>
        </section>
      )}
    </>
  );
}
