export const STATUS_CONFIG: Record<string, { label: string; dotClass: string; badgeClass: string }> = {
  active: {
    label: "Live",
    dotClass: "h-2 w-2 bg-emerald-400 shadow-[0_0_4px_oklch(0.7_0.18_160)]",
    badgeClass: "badge-live bg-emerald-500/10 text-emerald-700 border-emerald-500/25 dark:text-emerald-400 dark:border-emerald-500/20",
  },
  meeting_booked: {
    label: "Won",
    dotClass: "h-2 w-2 bg-primary",
    badgeClass: "bg-primary/10 text-primary border-primary/20",
  },
  terminated: {
    label: "Ended",
    dotClass: "h-2 w-2 bg-gray-400",
    badgeClass: "bg-gray-500/10 text-gray-600 border-gray-500/20 dark:text-gray-400",
  },
};

export const OUTREACH_CHANNELS = [
  { value: "email", label: "Email" },
  { value: "linkedin", label: "LinkedIn" },
  { value: "slack", label: "Slack" },
  { value: "phone", label: "Phone" },
  { value: "other", label: "Other" },
] as const;

export const VALID_SIGNAL_ACTIONS = [
  "outreach",
  "contact",
  "contact_remove",
  "agent",
  "booked",
] as const;

export type SignalAction = (typeof VALID_SIGNAL_ACTIONS)[number];
