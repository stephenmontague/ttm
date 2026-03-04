export function SiteFooter() {
  return (
    <footer className="border-t">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-6">
        <p className="text-xs text-muted-foreground">
          Powered by{" "}
          <a
            href="https://temporal.io"
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium underline-offset-4 hover:underline"
          >
            Temporal
          </a>
          {" "}durable execution
        </p>
        <p className="text-xs text-muted-foreground">
          Built with Next.js + Go
        </p>
      </div>
    </footer>
  );
}
