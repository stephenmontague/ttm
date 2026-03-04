import { notFound } from "next/navigation";
import Link from "next/link";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { ElapsedTimer } from "@/components/elapsed-timer";
import { StatusBadge } from "@/components/status-badge";
import { CompanyStats } from "@/components/company-stats";
import { ActivityFeed } from "@/components/activity-feed";
import { backendGet } from "@/lib/backend";
import { ArrowLeft } from "lucide-react";
import type { Company, ActivityFeedItem } from "@/lib/types";

interface PageProps {
  params: Promise<{ slug: string }>;
}

export default async function CompanyDetailPage({ params }: PageProps) {
  const { slug } = await params;

  const [companyRes, feedRes] = await Promise.all([
    backendGet<Company>(`/companies/${slug}`),
    backendGet<{ feed: ActivityFeedItem[] }>(`/companies/${slug}/feed`),
  ]);

  if (!companyRes.ok) {
    notFound();
  }

  const company = companyRes.data;
  const feed = feedRes.ok && feedRes.data.feed ? feedRes.data.feed : [];

  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader />
      <main className="flex-1">
        <div className="mx-auto max-w-4xl px-6 py-8">
          {/* Back link */}
          <Link
            href="/"
            className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to dashboard
          </Link>

          {/* Header */}
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-3xl font-bold tracking-tight">{company.CompanyName}</h1>
              <p className="mt-1 text-sm text-muted-foreground">
                Workflow started{" "}
                {new Date(company.StartedAt).toLocaleDateString("en-US", {
                  month: "long",
                  day: "numeric",
                  year: "numeric",
                })}
              </p>
            </div>
            <StatusBadge status={company.Status} />
          </div>

          {/* Elapsed Timer */}
          <div className="mt-8 rounded-xl border bg-card p-8">
            <p className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
              Time to Meeting
            </p>
            <ElapsedTimer startedAt={company.StartedAt} size="lg" />
          </div>

          {/* Stats */}
          <div className="mt-6">
            <CompanyStats
              outreachCount={company.OutreachCount}
              contactCount={company.ContactCount}
              restartCount={company.RestartCount}
              elapsedDays={company.ElapsedDays}
            />
          </div>

          {/* Activity Feed */}
          <div className="mt-8">
            <h2 className="mb-4 text-lg font-semibold tracking-tight">Activity</h2>
            <ActivityFeed feed={feed} />
          </div>
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
