import Link from "next/link";
import { listCompanies } from "@/lib/api";
import { WorkflowCard } from "@/components/workflow-card";
import { Company } from "@/lib/types";

export default async function Home() {
  let companies: Company[];
  try {
    companies = await listCompanies();
  } catch {
    companies = [];
  }

  const active = companies.filter((c) => c.Status !== "meeting_booked");
  const completed = companies.filter((c) => c.Status === "meeting_booked");

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">TTM Tracker</h1>
            <p className="text-sm text-muted-foreground">
              Temporal-powered outreach tracker
            </p>
          </div>
          <Link
            href="/admin"
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            Admin
          </Link>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <section>
          <h2 className="text-xl font-semibold mb-4">Active Workflows</h2>
          {active.length === 0 ? (
            <p className="text-muted-foreground">
              No active workflows. Start one from the{" "}
              <Link href="/admin" className="underline">
                admin panel
              </Link>
              .
            </p>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {active.map((company) => (
                <WorkflowCard key={company.ID} company={company} />
              ))}
            </div>
          )}
        </section>

        {completed.length > 0 && (
          <section className="mt-12">
            <h2 className="text-xl font-semibold mb-4">
              Wall of Wins
            </h2>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {completed.map((company) => (
                <WorkflowCard key={company.ID} company={company} />
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
