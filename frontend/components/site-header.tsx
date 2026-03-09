import Link from "next/link";
import { cookies } from "next/headers";
import { ThemeToggle } from "@/components/theme-toggle";
import { LogoutButton } from "@/components/logout-button";
import { Activity } from "lucide-react";

const COOKIE_NAME = process.env.SESSION_COOKIE_NAME || "ttm_session";

export async function SiteHeader() {
  const cookieStore = await cookies();
  const isLoggedIn = !!cookieStore.get(COOKIE_NAME)?.value;

  return (
    <header className="gradient-border-b sticky top-0 z-50 w-full bg-background/70 backdrop-blur-xl relative">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2.5 transition-opacity hover:opacity-80">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
            <Activity className="h-4 w-4 text-primary-foreground" />
          </div>
          <div className="flex flex-col">
            <span className="text-sm font-semibold leading-none tracking-tight">
              TTM Tracker
            </span>
            <span className="text-[10px] leading-none text-muted-foreground">
              Time to Meeting
            </span>
          </div>
        </Link>

        <nav className="flex items-center gap-1">
          <Link
            href="/"
            className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Dashboard
          </Link>
          {isLoggedIn && (
            <Link
              href="/admin"
              className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
            >
              Admin
            </Link>
          )}
          <div className="ml-2 h-4 w-px bg-border" />
          <ThemeToggle />
          {isLoggedIn && <LogoutButton />}
        </nav>
      </div>
    </header>
  );
}
