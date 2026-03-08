export interface Company {
  ID: string;
  CompanyName: string;
  Slug: string;
  StartedAt: string;
  Status: "active" | "meeting_booked" | "terminated" | "completed" | "canceled" | "failed";
  ElapsedDays: number;
  OutreachCount: number;
  ContactCount: number;
  RestartCount: number;
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

export interface Contact {
  Name: string;
  Role: string;
  LinkedIn: string;
  Active: boolean;
  AddedAt: string;
}

export interface WorkflowState {
  CompanyName: string;
  Slug: string;
  StartedAt: string;
  Status: string;
  CurrentContact: Contact | null;
  Contacts: Contact[];
  OutreachAttempts: {
    Timestamp: string;
    Channel: string;
    Notes: string;
    Contact: string;
  }[];
  AgentSuggestions: {
    Timestamp: string;
    TaskType: string;
    ContactName: string;
    Request: string;
    Response: string;
    DraftMessage: string;
  }[];
  AgentTaskInProgress: boolean;
  WorkerRestartCount: number;
  LastSnapshotAt: string;
  MeetingBookedAt: string | null;
  MeetingNotes: string;
}
