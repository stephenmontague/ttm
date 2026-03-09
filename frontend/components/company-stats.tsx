import { Mail, Users, RefreshCw, Clock } from "lucide-react";

interface CompanyStatsProps {
  outreachCount: number;
  contactCount: number;
  restartCount: number;
  elapsedDays: number;
}

export function CompanyStats({ outreachCount, contactCount, restartCount, elapsedDays }: CompanyStatsProps) {
  const stats = [
    {
      label: "Outreach Attempts",
      value: outreachCount,
      icon: Mail,
    },
    {
      label: "Contacts",
      value: contactCount,
      icon: Users,
    },
    {
      label: "Worker Restarts",
      value: restartCount,
      icon: RefreshCw,
    },
    {
      label: "Days Elapsed",
      value: elapsedDays,
      icon: Clock,
    },
  ];

  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 divide-x rounded-lg border bg-card">
      {stats.map((stat) => (
        <div key={stat.label} className="flex flex-col items-center gap-1 px-3 py-3 sm:px-6 sm:py-4">
          <stat.icon className="h-4 w-4 text-primary" />
          <span className="text-2xl font-semibold tabular-nums">{stat.value ?? 0}</span>
          <span className="text-xs text-muted-foreground">{stat.label}</span>
        </div>
      ))}
    </div>
  );
}
