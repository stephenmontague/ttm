import Link from "next/link";
import { getCompany, getCompanyFeed } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ElapsedTimer } from "@/components/elapsed-timer";

export default async function CompanyPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  let company;
  let feed;
  try {
    [company, feed] = await Promise.all([
      getCompany(slug),
      getCompanyFeed(slug),
    ]);
  } catch {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-muted-foreground">Company not found.</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-6">
          <Link
            href="/"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            &larr; Back
          </Link>
          <div className="flex items-center gap-3 mt-2">
            <h1 className="text-2xl font-bold">{company.CompanyName}</h1>
            <Badge variant="outline">{company.Status}</Badge>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 space-y-8">
        <Card>
          <CardHeader>
            <CardTitle>Time to Meeting</CardTitle>
          </CardHeader>
          <CardContent>
            <ElapsedTimer startedAt={company.StartedAt} />
            <div className="flex gap-6 mt-4 text-sm text-muted-foreground">
              <span>{company.OutreachCount} outreach attempts</span>
              {company.RestartCount > 0 && (
                <span>Workflow survived {company.RestartCount} restarts</span>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Activity Feed</CardTitle>
          </CardHeader>
          <CardContent>
            {feed.length === 0 ? (
              <p className="text-muted-foreground text-sm">
                No activity yet.
              </p>
            ) : (
              <ul className="space-y-3">
                {feed.map((item) => (
                  <li key={item.ID} className="flex gap-3 text-sm">
                    <span className="text-muted-foreground whitespace-nowrap">
                      {new Date(item.Timestamp).toLocaleDateString()}
                    </span>
                    <span>{item.Description}</span>
                    {item.Channel && (
                      <Badge variant="secondary">{item.Channel}</Badge>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
