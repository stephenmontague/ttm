import { Company, ActivityFeedItem, WorkflowState } from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8181/api";

async function fetchJSON<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API error ${res.status}: ${body}`);
  }

  return res.json();
}

// --- Public ---

export async function listCompanies(): Promise<Company[]> {
  const data = await fetchJSON<{ companies: Company[]; total: number }>(
    "/companies"
  );
  return data.companies;
}

export async function getCompany(slug: string): Promise<Company> {
  return fetchJSON<Company>(`/companies/${slug}`);
}

export async function getCompanyFeed(
  slug: string
): Promise<ActivityFeedItem[]> {
  const data = await fetchJSON<{ feed: ActivityFeedItem[]; total: number }>(
    `/companies/${slug}/feed`
  );
  return data.feed;
}

// --- Admin ---

export async function createCompany(
  companyName: string
): Promise<{ workflowId: string; slug: string }> {
  return fetchJSON("/admin/companies", {
    method: "POST",
    body: JSON.stringify({ companyName }),
  });
}

export async function getAdminCompany(slug: string): Promise<WorkflowState> {
  return fetchJSON<WorkflowState>(`/admin/companies/${slug}`);
}

export async function signalOutreach(
  slug: string,
  channel: string,
  notes: string
) {
  return fetchJSON(`/admin/companies/${slug}/signal/outreach`, {
    method: "POST",
    body: JSON.stringify({ channel, notes }),
  });
}

export async function signalUpdateContact(
  slug: string,
  name: string,
  role: string,
  linkedin: string
) {
  return fetchJSON(`/admin/companies/${slug}/signal/contact`, {
    method: "POST",
    body: JSON.stringify({ name, role, linkedin }),
  });
}

export async function signalMeetingBooked(
  slug: string,
  date: string,
  notes: string
) {
  return fetchJSON(`/admin/companies/${slug}/signal/booked`, {
    method: "POST",
    body: JSON.stringify({ date, notes }),
  });
}

export async function signalPause(slug: string) {
  return fetchJSON(`/admin/companies/${slug}/signal/pause`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}

export async function signalResume(slug: string) {
  return fetchJSON(`/admin/companies/${slug}/signal/resume`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}
