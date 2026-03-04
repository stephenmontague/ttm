"use client";

import { useCallback } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { usePolling } from "@/hooks/use-polling";
import { useSignal } from "@/hooks/use-signal";
import { ElapsedTimer } from "@/components/elapsed-timer";
import { StatusBadge } from "@/components/status-badge";
import { SignalPanel } from "@/components/signal-panel";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { ArrowLeft, RefreshCw, Users, Mail, X } from "lucide-react";
import type { WorkflowState, Contact } from "@/lib/types";

export default function AdminCompanyDetailPage() {
  const { slug } = useParams<{ slug: string }>();

  const fetcher = useCallback(async (): Promise<WorkflowState> => {
    const res = await fetch(`/api/admin/companies/${slug}`);
    if (!res.ok) throw new Error("Failed to fetch workflow state");
    return res.json();
  }, [slug]);

  const { data: state, loading, refresh } = usePolling<WorkflowState>({
    fetcher,
    interval: 10000,
  });

  const { send } = useSignal({ slug, onSuccess: refresh });

  if (loading && !state) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-8">
        <Skeleton className="h-8 w-48 mb-4" />
        <Skeleton className="h-32 w-full mb-4" />
        <div className="grid gap-4 lg:grid-cols-2">
          <Skeleton className="h-96" />
          <Skeleton className="h-96" />
        </div>
      </div>
    );
  }

  if (!state) {
    return (
      <div className="mx-auto max-w-6xl px-6 py-8">
        <p className="text-muted-foreground">Workflow not found.</p>
      </div>
    );
  }

  // Derive contacts list with backward compat for old CurrentContact.
  const contacts: Contact[] =
    state.Contacts && state.Contacts.length > 0
      ? state.Contacts
      : state.CurrentContact
        ? [{ ...state.CurrentContact, Active: true, AddedAt: "" }]
        : [];
  const activeContacts = contacts.filter((c) => c.Active);

  const handleRemoveContact = async (name: string) => {
    await send("contact_remove", { name });
  };

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      {/* Back link */}
      <Link
        href="/admin"
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to workflows
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{state.CompanyName}</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Started{" "}
            {new Date(state.StartedAt).toLocaleDateString("en-US", {
              month: "long",
              day: "numeric",
              year: "numeric",
            })}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <StatusBadge status={state.Status} />
          <Button variant="outline" size="icon" className="h-9 w-9" onClick={refresh}>
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Elapsed Timer */}
      <div className="mt-6 rounded-xl border bg-card p-8">
        <p className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
          Time to Meeting
        </p>
        <ElapsedTimer startedAt={state.StartedAt} size="lg" />
      </div>

      {/* Two-column layout */}
      <div className="mt-8 grid gap-8 lg:grid-cols-[1fr_380px]">
        {/* Left column — State & History */}
        <div className="space-y-6">
          {/* Contacts */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <Users className="h-4 w-4" />
                Contacts
                {activeContacts.length > 0 && (
                  <span className="ml-auto text-sm font-normal text-muted-foreground">
                    {activeContacts.length} active
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {contacts.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No contacts added. Use the signal panel to add one.
                </p>
              ) : (
                <div className="space-y-0">
                  {contacts.map((contact, i) => {
                    const outreachCount = (state.OutreachAttempts || []).filter(
                      (a) => a.Contact === contact.Name
                    ).length;
                    return (
                      <div key={contact.Name}>
                        {i > 0 && <Separator className="my-3" />}
                        <div className="flex items-start justify-between">
                          <div className="space-y-1">
                            <div className="flex items-center gap-2">
                              <p className="font-medium">{contact.Name}</p>
                              {!contact.Active && (
                                <Badge variant="outline" className="text-xs text-muted-foreground">
                                  Inactive
                                </Badge>
                              )}
                            </div>
                            <p className="text-sm text-muted-foreground">{contact.Role}</p>
                            {contact.LinkedIn && (
                              <a
                                href={contact.LinkedIn}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-sm text-blue-600 hover:underline dark:text-blue-400"
                              >
                                LinkedIn Profile
                              </a>
                            )}
                            <p className="text-xs text-muted-foreground">
                              {outreachCount} outreach{outreachCount !== 1 ? "es" : ""}
                            </p>
                          </div>
                          {contact.Active && state.Status === "active" && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 w-8 p-0"
                              onClick={() => handleRemoveContact(contact.Name)}
                            >
                              <X className="h-3.5 w-3.5" />
                            </Button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Outreach History */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <Mail className="h-4 w-4" />
                Outreach History
                {state.OutreachAttempts && state.OutreachAttempts.length > 0 && (
                  <span className="ml-auto text-sm font-normal text-muted-foreground">
                    {state.OutreachAttempts.length} total
                  </span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {!state.OutreachAttempts || state.OutreachAttempts.length === 0 ? (
                <p className="text-sm text-muted-foreground">No outreach attempts logged yet.</p>
              ) : (
                <div className="space-y-0">
                  {state.OutreachAttempts.slice().reverse().map((attempt, i) => (
                    <div key={i}>
                      {i > 0 && <Separator className="my-3" />}
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Badge variant="secondary" className="text-xs">
                            {attempt.Channel}
                          </Badge>
                          {attempt.Contact && (
                            <span className="text-xs text-muted-foreground">
                              to {attempt.Contact}
                            </span>
                          )}
                        </div>
                        <p className="text-sm">{attempt.Notes}</p>
                        <time className="text-xs text-muted-foreground">
                          {new Date(attempt.Timestamp).toLocaleDateString("en-US", {
                            month: "short",
                            day: "numeric",
                            hour: "numeric",
                            minute: "2-digit",
                          })}
                        </time>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Agent Suggestions (Phase 3) */}
          {state.AgentSuggestions && state.AgentSuggestions.length > 0 && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-base">Agent Suggestions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {state.AgentSuggestions.map((suggestion, i) => (
                    <div key={i} className="rounded-lg bg-muted/50 p-4">
                      <div className="flex items-center gap-2 mb-2">
                        <Badge variant="outline" className="text-xs">
                          {suggestion.TaskType}
                        </Badge>
                        <time className="text-xs text-muted-foreground">
                          {new Date(suggestion.Timestamp).toLocaleDateString()}
                        </time>
                      </div>
                      <p className="text-sm whitespace-pre-wrap">{suggestion.Response}</p>
                      {suggestion.DraftMessage && (
                        <div className="mt-3 rounded-md border bg-card p-3">
                          <p className="mb-1 text-xs font-medium text-muted-foreground">
                            Draft Message
                          </p>
                          <p className="text-sm whitespace-pre-wrap">{suggestion.DraftMessage}</p>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Right column — Signal Panel */}
        <div>
          <h2 className="mb-4 text-sm font-medium uppercase tracking-wider text-muted-foreground">
            Signals
          </h2>
          <SignalPanel slug={slug} status={state.Status} contacts={activeContacts} onSuccess={refresh} />
        </div>
      </div>
    </div>
  );
}
