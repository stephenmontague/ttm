import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { ElapsedTimer } from "@/components/elapsed-timer";
import { StatusBadge } from "@/components/status-badge";
import { Mail, Users } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Company } from "@/lib/types";

interface WorkflowCardProps {
  company: Company;
  index?: number;
}

export function WorkflowCard({ company, index = 0 }: WorkflowCardProps) {
  return (
    <Link href={`/company/${company.Slug}`}>
      <Card
        className={cn(
          "group relative overflow-hidden transition-all duration-300",
          "hover:shadow-[0_8px_30px_var(--glow-primary)] hover:-translate-y-1",
          "animate-card-enter",
          company.Status === "active" && "border-l-2 border-l-primary"
        )}
        style={{ animationDelay: `${index * 80}ms` }}
      >
        <div className="absolute inset-0 bg-linear-to-br from-primary/0 to-primary/0 transition-all duration-300 group-hover:from-primary/3 group-hover:to-transparent" />
        <CardContent className="relative p-6">
          <div className="flex items-start justify-between gap-4">
            <h3 className="font-semibold leading-none tracking-tight">
              {company.CompanyName}
            </h3>
            <StatusBadge status={company.Status} />
          </div>

          <div className="mt-4 -mx-6 px-6 py-2 bg-muted/30 rounded-sm">
            <ElapsedTimer startedAt={company.StartedAt} size="sm" className="text-base" />
          </div>

          <div className="mt-4 flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-1.5">
              <Mail className="h-3.5 w-3.5" />
              <span>{company.OutreachCount} outreach{company.OutreachCount !== 1 ? "es" : ""}</span>
            </div>
            {company.ContactCount > 0 && (
              <div className="flex items-center gap-1.5">
                <Users className="h-3.5 w-3.5" />
                <span>{company.ContactCount} contact{company.ContactCount !== 1 ? "s" : ""}</span>
              </div>
            )}
          </div>

          {company.RestartCount > 0 && (
            <div className="mt-3 rounded-md bg-muted/50 px-3 py-1.5 text-xs text-muted-foreground">
              Workflow survived {company.RestartCount} restart{company.RestartCount !== 1 ? "s" : ""}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
