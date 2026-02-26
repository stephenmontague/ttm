export interface Company {
  ID: string;
  CompanyName: string;
  Slug: string;
  StartedAt: string;
  Status: "active" | "paused" | "meeting_booked";
  ElapsedDays: number;
  OutreachCount: number;
  RestartCount: number;
  CurrentContactRole: string | null;
  MeetingBookedAt: string | null;
  LastSnapshotAt: string | null;
  UpdatedAt: string;
}

export interface ActivityFeedItem {
  ID: number;
  WorkflowID: string;
  Timestamp: string;
  EventType: string;
  Description: string;
  Channel: string | null;
  CreatedAt: string;
}

// Full workflow state returned by the admin query endpoint
export interface WorkflowState {
  CompanyName: string;
  Slug: string;
  StartedAt: string;
  Status: string;
  CurrentContact: {
    Name: string;
    Role: string;
    LinkedIn: string;
  } | null;
  OutreachAttempts: {
    Timestamp: string;
    Channel: string;
    Notes: string;
    Contact: string;
  }[];
  AgentSuggestions: {
    Timestamp: string;
    TaskType: string;
    Request: string;
    Response: string;
    DraftMessage: string;
  }[];
  WorkerRestartCount: number;
  LastSnapshotAt: string;
  MeetingBookedAt: string | null;
  MeetingNotes: string;
}
