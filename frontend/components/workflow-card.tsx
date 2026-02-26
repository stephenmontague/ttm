import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ElapsedTimer } from "@/components/elapsed-timer";
import { Company } from "@/lib/types";

interface WorkflowCardProps {
  company: Company;
}

const statusColors: Record<string, string> = {
  active: "bg-green-500/10 text-green-500 border-green-500/20",
  paused: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
  meeting_booked: "bg-blue-500/10 text-blue-500 border-blue-500/20",
};

export function WorkflowCard({ company }: WorkflowCardProps) {
  return (
    <Link href={`/company/${company.Slug}`}>
      <Card className="hover:border-primary/50 transition-colors cursor-pointer">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-lg">{company.CompanyName}</CardTitle>
          <Badge
            variant="outline"
            className={statusColors[company.Status] || ""}
          >
            {company.Status === "meeting_booked"
              ? "Meeting Booked!"
              : company.Status}
          </Badge>
        </CardHeader>
        <CardContent>
          <ElapsedTimer startedAt={company.StartedAt} />
          <div className="flex gap-4 mt-3 text-sm text-muted-foreground">
            <span>{company.OutreachCount} outreach attempts</span>
            {company.RestartCount > 0 && (
              <span>Survived {company.RestartCount} restarts</span>
            )}
          </div>
          {company.CurrentContactRole && (
            <p className="text-sm text-muted-foreground mt-1">
              Current contact: {company.CurrentContactRole}
            </p>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
