import { Badge } from "@/components/ui/badge";
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
      <div className="absolute left-2.75 top-2 bottom-2 w-px bg-border" />
      {feed.map((item) => (
        <div key={item.ID} className="relative flex gap-4 py-3">
          <div className="relative z-10 mt-1.5 h-2.25 w-2.25 shrink-0 rounded-full border-2 border-border bg-background" />
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
