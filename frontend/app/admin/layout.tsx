import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader />
      <div className="border-b bg-amber-500/5">
        <div className="mx-auto max-w-6xl px-6 py-2">
          <p className="text-xs font-medium text-amber-700 dark:text-amber-400">
            Admin Panel
          </p>
        </div>
      </div>
      <main className="flex-1">{children}</main>
      <SiteFooter />
    </div>
  );
}
