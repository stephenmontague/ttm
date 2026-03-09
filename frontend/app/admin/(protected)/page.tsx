"use client";

import { useCallback } from "react";
import Link from "next/link";
import { usePolling } from "@/hooks/use-polling";
import { CreateWorkflowForm } from "@/components/create-workflow-form";
import { StatusBadge } from "@/components/status-badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Activity, Trophy, ChevronRight, AlertCircle } from "lucide-react";
import type { Company } from "@/lib/types";

async function fetchCompanies(): Promise<Company[]> {
  const res = await fetch("/api/admin/companies");
  if (!res.ok) throw new Error(`Failed to load companies (${res.status})`);
  const data = await res.json();
  return data.companies || [];
}

export default function AdminPage() {
  const { data: companies, loading, error, refresh } = usePolling<Company[]>({
    fetcher: fetchCompanies,
    interval: 10000,
  });

  const handleCreated = useCallback(() => {
    setTimeout(refresh, 1000);
  }, [refresh]);

  const list = companies ?? [];
  const activeCount = list.filter((c) => c.Status === "active").length;
  const wonCount = list.filter((c) => c.Status === "meeting_booked").length;

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      <h1 className="text-3xl font-bold tracking-tight">Workflows</h1>
      <p className="mt-1 text-muted-foreground">
        Manage outreach workflows and send signals
      </p>

      {/* Error banner */}
      {error && (
        <div className="mt-6 flex items-center gap-2 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
          <AlertCircle className="h-4 w-4 shrink-0" />
          {error}. Retrying...
        </div>
      )}

      {/* Stats */}
      <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Card>
          <CardContent className="flex items-center gap-3 p-4">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
              <Activity className="h-4 w-4" />
            </div>
            <div>
              <p className="text-2xl font-semibold tabular-nums">{list.length}</p>
              <p className="text-xs text-muted-foreground">Total</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 p-4">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-500/10">
              <Activity className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
            </div>
            <div>
              <p className="text-2xl font-semibold tabular-nums">{activeCount}</p>
              <p className="text-xs text-muted-foreground">Active</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 p-4">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-500/10">
              <Trophy className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <p className="text-2xl font-semibold tabular-nums">{wonCount}</p>
              <p className="text-xs text-muted-foreground">Won</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 p-4">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
              <Activity className="h-4 w-4 text-muted-foreground" />
            </div>
            <div>
              <p className="text-2xl font-semibold tabular-nums">
                {list.reduce((sum, c) => sum + c.OutreachCount, 0)}
              </p>
              <p className="text-xs text-muted-foreground">Outreaches</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Create */}
      <Card className="mt-8">
        <CardHeader>
          <CardTitle>Start New Workflow</CardTitle>
          <CardDescription>
            Create a new outreach workflow for a target company
          </CardDescription>
        </CardHeader>
        <CardContent>
          <CreateWorkflowForm onCreated={handleCreated} />
        </CardContent>
      </Card>

      {/* Company List */}
      <div className="mt-8">
        <h2 className="mb-4 text-lg font-semibold tracking-tight">All Workflows</h2>
        {loading && list.length === 0 ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-16 rounded-lg" />
            ))}
          </div>
        ) : list.length === 0 ? (
          <div className="rounded-xl border border-dashed p-12 text-center">
            <p className="text-sm text-muted-foreground">
              No workflows yet. Create one above to get started.
            </p>
          </div>
        ) : (
          <div className="rounded-xl border">
            {list.map((company, i) => (
              <Link
                key={company.ID}
                href={`/admin/company/${company.Slug}`}
                className={`flex items-center justify-between px-4 py-3 transition-colors hover:bg-muted/50 ${
                  i !== list.length - 1 ? "border-b" : ""
                }`}
              >
                <div className="flex items-center gap-4">
                  <div className="min-w-0">
                    <p className="font-medium truncate">{company.CompanyName}</p>
                    <p className="text-xs text-muted-foreground">
                      {company.OutreachCount} outreach{company.OutreachCount !== 1 ? "es" : ""} &middot;{" "}
                      {company.ElapsedDays} day{company.ElapsedDays !== 1 ? "s" : ""}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <StatusBadge status={company.Status} />
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
