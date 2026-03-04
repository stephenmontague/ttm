import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { ElapsedTimer } from "@/components/elapsed-timer";
import { StatusBadge } from "@/components/status-badge";
import { Mail, Users } from "lucide-react";
import type { Company } from "@/lib/types";

interface WorkflowCardProps {
  company: Company;
}

export function WorkflowCard({ company }: WorkflowCardProps) {
  return (
    <Link href={`/company/${company.Slug}`}>
      <Card className="group relative overflow-hidden transition-all duration-200 hover:shadow-lg hover:-translate-y-0.5">
        <CardContent className="p-6">
          <div className="flex items-start justify-between gap-4">
            <h3 className="font-semibold leading-none tracking-tight">
              {company.CompanyName}
            </h3>
            <StatusBadge status={company.Status} />
          </div>

          <div className="mt-4">
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
