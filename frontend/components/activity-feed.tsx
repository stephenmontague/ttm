import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { ActivityFeedItem } from "@/lib/types";

interface ActivityFeedProps {
  feed: ActivityFeedItem[];
}

export function ActivityFeed({ feed }: ActivityFeedProps) {
  if (feed.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-sm text-muted-foreground">No activity yet</p>
      </div>
    );
  }

  return (
    <div className="relative space-y-0">
      <div className="absolute left-3 top-3 bottom-3 w-px bg-linear-to-b from-primary/50 via-primary/20 to-transparent" />
      {feed.map((item, index) => (
        <div
          key={item.ID}
          className="animate-fade-in relative flex gap-4 py-3"
          style={{ animationDelay: `${index * 50}ms` }}
        >
          <div
            className={cn(
              "relative z-10 mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full border-2",
              index === 0
                ? "border-primary bg-primary shadow-[0_0_6px_var(--glow-primary)]"
                : "border-border bg-background"
            )}
          />
          <div className="flex-1 space-y-1">
            <p className="text-sm leading-relaxed">{item.Description}</p>
            <div className="flex items-center gap-2">
              <time className="text-xs text-muted-foreground">
                {new Date(item.Timestamp).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                  hour: "numeric",
                  minute: "2-digit",
                })}
              </time>
              {item.Channel && (
                <Badge variant="secondary" className="h-5 px-1.5 text-[10px]">
                  {item.Channel}
                </Badge>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
