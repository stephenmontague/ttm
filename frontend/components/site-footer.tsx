export function SiteFooter() {
  return (
    <footer className="gradient-border-t relative">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <div className="flex items-center gap-2">
          <div className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse-live" />
          <p className="text-xs text-muted-foreground">
            Powered by{" "}
            <a
              href="https://temporal.io"
              target="_blank"
              rel="noopener noreferrer"
              className="font-semibold text-foreground transition-colors hover:text-primary"
            >
              Temporal
            </a>
            {" "}durable execution
          </p>
        </div>
      </div>
    </footer>
  );
}
