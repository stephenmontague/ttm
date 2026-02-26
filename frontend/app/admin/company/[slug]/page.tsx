"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ElapsedTimer } from "@/components/elapsed-timer";
import {
  getAdminCompany,
  signalOutreach,
  signalUpdateContact,
  signalMeetingBooked,
  signalPause,
  signalResume,
} from "@/lib/api";
import { WorkflowState } from "@/lib/types";

export default function AdminCompanyPage() {
  const params = useParams<{ slug: string }>();
  const slug = params.slug;

  const [state, setState] = useState<WorkflowState | null>(null);
  const [error, setError] = useState("");
  const [signalStatus, setSignalStatus] = useState("");

  // Outreach form
  const [outreachChannel, setOutreachChannel] = useState("email");
  const [outreachNotes, setOutreachNotes] = useState("");

  // Contact form
  const [contactName, setContactName] = useState("");
  const [contactRole, setContactRole] = useState("");
  const [contactLinkedIn, setContactLinkedIn] = useState("");

  // Meeting form
  const [meetingNotes, setMeetingNotes] = useState("");

  const loadState = useCallback(async () => {
    try {
      const data = await getAdminCompany(slug);
      setState(data);
      setError("");
    } catch (err) {
      setError("Failed to load workflow state");
    }
  }, [slug]);

  useEffect(() => {
    loadState();
  }, [loadState]);

  async function handleSignal(name: string, fn: () => Promise<unknown>) {
    setSignalStatus(`Sending ${name}...`);
    try {
      await fn();
      setSignalStatus(`${name} sent!`);
      setTimeout(loadState, 500);
    } catch (err) {
      setSignalStatus(
        `Failed: ${err instanceof Error ? err.message : "Unknown error"}`
      );
    }
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-destructive">{error}</p>
      </div>
    );
  }

  if (!state) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-6">
          <Link
            href="/admin"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            &larr; Back to Admin
          </Link>
          <div className="flex items-center gap-3 mt-2">
            <h1 className="text-2xl font-bold">{state.CompanyName}</h1>
            <Badge variant="outline">{state.Status}</Badge>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 space-y-6">
        {signalStatus && (
          <div className="bg-muted px-4 py-2 rounded-md text-sm">
            {signalStatus}
          </div>
        )}

        {/* Live Timer */}
        <Card>
          <CardHeader>
            <CardTitle>Elapsed Time</CardTitle>
          </CardHeader>
          <CardContent>
            <ElapsedTimer startedAt={state.StartedAt} />
          </CardContent>
        </Card>

        {/* Pause / Resume */}
        <div className="flex gap-2">
          {state.Status === "active" ? (
            <Button
              variant="outline"
              onClick={() =>
                handleSignal("pause", () => signalPause(slug))
              }
            >
              Pause Workflow
            </Button>
          ) : state.Status === "paused" ? (
            <Button
              onClick={() =>
                handleSignal("resume", () => signalResume(slug))
              }
            >
              Resume Workflow
            </Button>
          ) : null}
          <Button variant="outline" onClick={loadState}>
            Refresh State
          </Button>
        </div>

        {/* Current Contact */}
        <Card>
          <CardHeader>
            <CardTitle>Current Contact</CardTitle>
          </CardHeader>
          <CardContent>
            {state.CurrentContact ? (
              <div className="text-sm space-y-1">
                <p>
                  <span className="font-medium">{state.CurrentContact.Name}</span>{" "}
                  &mdash; {state.CurrentContact.Role}
                </p>
                {state.CurrentContact.LinkedIn && (
                  <p className="text-muted-foreground">
                    {state.CurrentContact.LinkedIn}
                  </p>
                )}
              </div>
            ) : (
              <p className="text-muted-foreground text-sm">No contact set</p>
            )}

            <div className="grid grid-cols-3 gap-2 mt-4">
              <div>
                <Label>Name</Label>
                <Input
                  value={contactName}
                  onChange={(e) => setContactName(e.target.value)}
                  placeholder="Jane Doe"
                />
              </div>
              <div>
                <Label>Role</Label>
                <Input
                  value={contactRole}
                  onChange={(e) => setContactRole(e.target.value)}
                  placeholder="VP Engineering"
                />
              </div>
              <div>
                <Label>LinkedIn</Label>
                <Input
                  value={contactLinkedIn}
                  onChange={(e) => setContactLinkedIn(e.target.value)}
                  placeholder="linkedin.com/in/..."
                />
              </div>
            </div>
            <Button
              className="mt-2"
              size="sm"
              onClick={() =>
                handleSignal("update contact", () =>
                  signalUpdateContact(
                    slug,
                    contactName,
                    contactRole,
                    contactLinkedIn
                  )
                )
              }
            >
              Update Contact
            </Button>
          </CardContent>
        </Card>

        {/* Log Outreach */}
        <Card>
          <CardHeader>
            <CardTitle>Log Outreach</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <Select value={outreachChannel} onValueChange={setOutreachChannel}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="email">Email</SelectItem>
                  <SelectItem value="linkedin">LinkedIn</SelectItem>
                  <SelectItem value="slack">Slack</SelectItem>
                  <SelectItem value="phone">Phone</SelectItem>
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
              <Textarea
                value={outreachNotes}
                onChange={(e) => setOutreachNotes(e.target.value)}
                placeholder="Notes about the outreach..."
                className="flex-1"
                rows={2}
              />
            </div>
            <Button
              className="mt-2"
              size="sm"
              onClick={() =>
                handleSignal("outreach", () =>
                  signalOutreach(slug, outreachChannel, outreachNotes)
                )
              }
            >
              Log Outreach
            </Button>
          </CardContent>
        </Card>

        {/* Meeting Booked */}
        <Card>
          <CardHeader>
            <CardTitle>Meeting Booked</CardTitle>
          </CardHeader>
          <CardContent>
            <Textarea
              value={meetingNotes}
              onChange={(e) => setMeetingNotes(e.target.value)}
              placeholder="Meeting details..."
              rows={2}
            />
            <Button
              className="mt-2"
              size="sm"
              variant="default"
              onClick={() =>
                handleSignal("meeting booked", () =>
                  signalMeetingBooked(slug, new Date().toISOString(), meetingNotes)
                )
              }
            >
              Mark Meeting Booked
            </Button>
          </CardContent>
        </Card>

        {/* Outreach History */}
        <Card>
          <CardHeader>
            <CardTitle>Outreach History</CardTitle>
          </CardHeader>
          <CardContent>
            {state.OutreachAttempts.length === 0 ? (
              <p className="text-muted-foreground text-sm">
                No outreach logged yet.
              </p>
            ) : (
              <ul className="space-y-2">
                {state.OutreachAttempts.map((attempt, i) => (
                  <li key={i} className="flex gap-3 text-sm">
                    <span className="text-muted-foreground whitespace-nowrap">
                      {new Date(attempt.Timestamp).toLocaleString()}
                    </span>
                    <Badge variant="secondary">{attempt.Channel}</Badge>
                    {attempt.Contact && (
                      <span className="text-muted-foreground">
                        to {attempt.Contact}
                      </span>
                    )}
                    <span>{attempt.Notes}</span>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
