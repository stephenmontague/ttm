export const STATUS_CONFIG: Record<string, { label: string; dotClass: string; badgeClass: string }> = {
  active: {
    label: "Live",
    dotClass: "bg-emerald-500 animate-pulse",
    badgeClass: "bg-emerald-500/10 text-emerald-700 border-emerald-500/20 dark:text-emerald-400",
  },
  meeting_booked: {
    label: "Won",
    dotClass: "bg-blue-500",
    badgeClass: "bg-blue-500/10 text-blue-700 border-blue-500/20 dark:text-blue-400",
  },
  terminated: {
    label: "Ended",
    dotClass: "bg-gray-400",
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
