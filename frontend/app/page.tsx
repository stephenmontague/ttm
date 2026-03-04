import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { backendGet } from "@/lib/backend";
import type { Company } from "@/lib/types";
import { DashboardContent } from "./dashboard-content";

async function getCompanies(): Promise<Company[]> {
  const { data, ok } = await backendGet<{ companies: Company[] }>("/companies");
  if (!ok || !data.companies) return [];
  return data.companies;
}

export default async function HomePage() {
  const companies = await getCompanies();

  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader />
      <main className="flex-1">
        <DashboardContent initialCompanies={companies} />
      </main>
      <SiteFooter />
    </div>
  );
}
