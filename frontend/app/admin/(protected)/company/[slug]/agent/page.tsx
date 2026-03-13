"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import ReactMarkdown from "react-markdown";
import { usePolling } from "@/hooks/use-polling";
import { StatusBadge } from "@/components/status-badge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  ArrowLeft,
  Bot,
  Copy,
  Check,
  Loader2,
  Mail,
  Send,
  Users,
} from "lucide-react";
import { toast } from "sonner";
import type { WorkflowState, Contact } from "@/lib/types";

export default function AgentPage() {
  const { slug } = useParams<{ slug: string }>();

  // Track whether agent is working to adapt poll interval
  const [agentPending, setAgentPending] = useState(false);
  const prevSuggestionCount = useRef<number | null>(null);

  const fetcher = useCallback(async (): Promise<WorkflowState> => {
    const res = await fetch(`/api/admin/companies/${slug}`);
    if (!res.ok) throw new Error("Failed to fetch workflow state");
    return res.json();
  }, [slug]);

  const pollInterval = agentPending ? 2000 : 10000;

  const { data: state, loading, refresh } = usePolling<WorkflowState>({
    fetcher,
    interval: pollInterval,
  });

  // Detect when agent finishes: new suggestion appears or flag clears
  useEffect(() => {
    if (!state) return;
    const count = state.AgentSuggestions?.length ?? 0;
    if (
      agentPending &&
      prevSuggestionCount.current !== null &&
      count > prevSuggestionCount.current
    ) {
      setAgentPending(false);
      toast.success("Agent suggestion ready");
    }
    // Also clear if backend flag is false and we were pending
    if (agentPending && !state.AgentTaskInProgress && prevSuggestionCount.current !== null && count > prevSuggestionCount.current) {
      setAgentPending(false);
    }
    prevSuggestionCount.current = count;
  }, [state, agentPending]);

  // Form state
  const [selectedContact, setSelectedContact] = useState("");
  const [taskType, setTaskType] = useState("");
  const [context, setContext] = useState("");
  const [sending, setSending] = useState(false);

  // Contact filter for suggestion history
  const [filterContact, setFilterContact] = useState<string | null>(null);

  // Copy state
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const handleCopy = async (text: string, index: number) => {
    await navigator.clipboard.writeText(text);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const handleSubmit = async () => {
    if (!taskType) return;
    setSending(true);
    try {
      const res = await fetch(`/api/admin/companies/${slug}/signal/agent`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          taskType,
          context,
          contactName: selectedContact === "__any__" ? undefined : selectedContact || undefined,
        }),
      });
      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Failed to send agent request");
      }
      toast.success("Agent request sent");
      setAgentPending(true);
      setContext("");
      // Refresh immediately to pick up the in-progress flag
      setTimeout(refresh, 500);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to send agent request"
      );
    } finally {
      setSending(false);
    }
  };

  // Quick action: pre-fill contact + task type
  const handleDraftFor = (contactName: string) => {
    setSelectedContact(contactName);
    setTaskType("draft_message");
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

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

  const contacts: Contact[] =
    state.Contacts && state.Contacts.length > 0
      ? state.Contacts
      : state.CurrentContact
        ? [{ ...state.CurrentContact, Active: true, AddedAt: "" }]
        : [];
  const activeContacts = contacts.filter((c) => c.Active);
  const isActive = state.Status === "active";
  const isAgentWorking = agentPending || state.AgentTaskInProgress;

  // Filter suggestions
  const suggestions = (state.AgentSuggestions || []).slice().reverse();
  const filteredSuggestions = filterContact
    ? suggestions.filter((s) => s.ContactName === filterContact)
    : suggestions;

  // Per-contact stats
  const contactStats = (name: string) => {
    const outreaches = (state.OutreachAttempts || []).filter(
      (a) => a.Contact === name
    ).length;
    const agentSugs = (state.AgentSuggestions || []).filter(
      (s) => s.ContactName === name
    ).length;
    return { outreaches, agentSugs };
  };

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      {/* Header */}
      <Link
        href={`/admin/company/${slug}`}
        className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to {state.CompanyName}
      </Link>

      <div className="flex items-start justify-between gap-4 mb-8">
        <div>
          <div className="flex items-center gap-3">
            <Bot className="h-6 w-6 text-purple-500" />
            <h1 className="text-2xl font-bold tracking-tight">
              Agent — {state.CompanyName}
            </h1>
          </div>
        </div>
        <StatusBadge status={state.Status} />
      </div>

      {/* Two-column layout */}
      <div className="grid gap-8 lg:grid-cols-[1fr_320px]">
        {/* Left column */}
        <div className="space-y-6">
          {/* Agent request form */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <Send className="h-4 w-4" />
                Ask the Agent
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div
                className={
                  !isActive ? "opacity-50 pointer-events-none space-y-4" : "space-y-4"
                }
              >
                <div className="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-4">
                  <div className="space-y-1.5 min-w-0">
                    <Label className="text-xs">Contact (optional)</Label>
                    <Select
                      value={selectedContact}
                      onValueChange={setSelectedContact}
                    >
                      <SelectTrigger className="truncate">
                        <SelectValue placeholder="Any / General" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__any__">Any / General</SelectItem>
                        {activeContacts.map((c) => (
                          <SelectItem key={c.Name} value={c.Name}>
                            {c.Name} ({c.Role})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5 w-48 shrink-0">
                    <Label className="text-xs">Task</Label>
                    <Select value={taskType} onValueChange={setTaskType}>
                      <SelectTrigger>
                        <SelectValue placeholder="What do you need?" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="draft_message">
                          Draft a message
                        </SelectItem>
                        <SelectItem value="suggest_contact">
                          Suggest a contact
                        </SelectItem>
                        <SelectItem value="next_action">
                          Recommend next action
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs">
                    Additional Context (optional)
                  </Label>
                  <Textarea
                    placeholder="Any specific focus or details..."
                    value={context}
                    onChange={(e) => setContext(e.target.value)}
                    rows={2}
                  />
                </div>
                <Button
                  className="w-full"
                  disabled={sending || isAgentWorking || !taskType}
                  onClick={handleSubmit}
                >
                  {sending || isAgentWorking ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Bot className="mr-2 h-4 w-4" />
                  )}
                  {isAgentWorking ? "Agent is thinking..." : "Ask Agent"}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Agent thinking indicator */}
          {isAgentWorking && (
            <Card className="border-purple-500/30 bg-purple-500/5">
              <CardContent className="py-6">
                <div className="flex items-center gap-3">
                  <Loader2 className="h-5 w-5 animate-spin text-purple-500" />
                  <div>
                    <p className="font-medium text-sm">Agent is working...</p>
                    <p className="text-xs text-muted-foreground">
                      This usually takes 10–60 seconds. The page will update
                      automatically.
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Suggestion history */}
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold tracking-tight">
                Suggestions
                {filteredSuggestions.length > 0 && (
                  <span className="ml-2 text-sm font-normal text-muted-foreground">
                    {filteredSuggestions.length}
                  </span>
                )}
              </h2>
              {filterContact && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setFilterContact(null)}
                  className="text-xs"
                >
                  Show all
                </Button>
              )}
            </div>

            {filteredSuggestions.length === 0 ? (
              <Card>
                <CardContent className="py-8 text-center">
                  <p className="text-sm text-muted-foreground">
                    {filterContact
                      ? `No suggestions for ${filterContact} yet.`
                      : "No agent suggestions yet. Ask the agent for help above."}
                  </p>
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-4">
                {filteredSuggestions.map((suggestion, i) => (
                  <Card key={i}>
                    <CardContent className="pt-5">
                      <div className="flex items-center gap-2 mb-3">
                        <Badge variant="outline" className="text-xs">
                          {suggestion.TaskType.replace("_", " ")}
                        </Badge>
                        {suggestion.ContactName && (
                          <Badge
                            variant="secondary"
                            className="text-xs cursor-pointer"
                            onClick={() =>
                              setFilterContact(suggestion.ContactName)
                            }
                          >
                            {suggestion.ContactName}
                          </Badge>
                        )}
                        <time className="ml-auto text-xs text-muted-foreground">
                          {new Date(suggestion.Timestamp).toLocaleDateString(
                            "en-US",
                            {
                              month: "short",
                              day: "numeric",
                              hour: "numeric",
                              minute: "2-digit",
                            }
                          )}
                        </time>
                      </div>
                      <div className="markdown">
                        <ReactMarkdown>{suggestion.Response}</ReactMarkdown>
                      </div>
                      {suggestion.DraftMessage && (
                        <div className="mt-4 rounded-md border bg-muted/30 p-4">
                          <div className="flex items-center justify-between mb-2">
                            <p className="text-xs font-medium text-muted-foreground">
                              Draft Message
                            </p>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 px-2 text-xs"
                              onClick={() =>
                                handleCopy(suggestion.DraftMessage, i)
                              }
                            >
                              {copiedIndex === i ? (
                                <>
                                  <Check className="mr-1 h-3 w-3" /> Copied
                                </>
                              ) : (
                                <>
                                  <Copy className="mr-1 h-3 w-3" /> Copy
                                </>
                              )}
                            </Button>
                          </div>
                          <div className="markdown">
                            <ReactMarkdown>{suggestion.DraftMessage}</ReactMarkdown>
                          </div>
                        </div>
                      )}
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right column — Contacts sidebar */}
        <div>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <Users className="h-4 w-4" />
                Contacts
              </CardTitle>
            </CardHeader>
            <CardContent>
              {contacts.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No contacts yet. Add one from the{" "}
                  <Link
                    href={`/admin/company/${slug}`}
                    className="underline underline-offset-4 hover:text-foreground"
                  >
                    company page
                  </Link>
                  .
                </p>
              ) : (
                <div className="space-y-0">
                  {contacts.map((contact, i) => {
                    const stats = contactStats(contact.Name);
                    const isFiltered = filterContact === contact.Name;
                    return (
                      <div key={contact.Name}>
                        {i > 0 && <Separator className="my-3" />}
                        <div
                          className={`rounded-md p-2 -mx-2 cursor-pointer transition-colors ${
                            isFiltered
                              ? "bg-purple-500/10"
                              : "hover:bg-muted/50"
                          }`}
                          onClick={() =>
                            setFilterContact(
                              isFiltered ? null : contact.Name
                            )
                          }
                        >
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="font-medium text-sm">
                                {contact.Name}
                              </p>
                              <p className="text-xs text-muted-foreground">
                                {contact.Role}
                              </p>
                            </div>
                            {!contact.Active && (
                              <Badge
                                variant="outline"
                                className="text-xs text-muted-foreground"
                              >
                                Inactive
                              </Badge>
                            )}
                          </div>
                          <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
                            <span className="flex items-center gap-1">
                              <Mail className="h-3 w-3" />
                              {stats.outreaches} outreach
                              {stats.outreaches !== 1 ? "es" : ""}
                            </span>
                            <span className="flex items-center gap-1">
                              <Bot className="h-3 w-3" />
                              {stats.agentSugs} suggestion
                              {stats.agentSugs !== 1 ? "s" : ""}
                            </span>
                          </div>
                          {contact.Active && isActive && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="mt-2 h-7 text-xs w-full justify-start text-purple-600 dark:text-purple-400 hover:text-purple-700"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDraftFor(contact.Name);
                              }}
                            >
                              <Send className="mr-1.5 h-3 w-3" />
                              Draft message for {contact.Name.split(" ")[0]}
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
        </div>
      </div>
    </div>
  );
}
