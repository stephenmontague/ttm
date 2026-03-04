"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSignal } from "@/hooks/use-signal";
import { OUTREACH_CHANNELS } from "@/lib/constants";
import { Mail, UserPlus, Trophy, Loader2 } from "lucide-react";
import type { Contact } from "@/lib/types";

interface SignalPanelProps {
  slug: string;
  status: string;
  contacts: Contact[];
  onSuccess: () => void;
}

export function SignalPanel({ slug, status, contacts, onSuccess }: SignalPanelProps) {
  const { send, loading } = useSignal({ slug, onSuccess });
  const isActive = status === "active";

  const [outreachContact, setOutreachContact] = useState("");
  const [outreachChannel, setOutreachChannel] = useState("");
  const [outreachNotes, setOutreachNotes] = useState("");

  const [contactName, setContactName] = useState("");
  const [contactRole, setContactRole] = useState("");
  const [contactLinkedin, setContactLinkedin] = useState("");

  const [meetingNotes, setMeetingNotes] = useState("");

  const handleOutreach = async () => {
    await send("outreach", { channel: outreachChannel, notes: outreachNotes, contactName: outreachContact });
    setOutreachContact("");
    setOutreachChannel("");
    setOutreachNotes("");
  };

  const handleContact = async () => {
    await send("contact", { name: contactName, role: contactRole, linkedin: contactLinkedin });
    setContactName("");
    setContactRole("");
    setContactLinkedin("");
  };

  const handleMeetingBooked = async () => {
    await send("booked", { date: new Date().toISOString(), notes: meetingNotes });
    setMeetingNotes("");
  };

  const activeContacts = contacts.filter((c) => c.Active);

  return (
    <div className="space-y-4">
      {/* Log Outreach */}
      <Card className={!isActive ? "opacity-50 pointer-events-none" : ""}>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Mail className="h-4 w-4" />
            Log Outreach
          </CardTitle>
          <CardDescription>Record an outreach attempt</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-1.5">
            <Label htmlFor="outreach-contact" className="text-xs">Contact</Label>
            <Select value={outreachContact} onValueChange={setOutreachContact}>
              <SelectTrigger id="outreach-contact">
                <SelectValue placeholder="Select contact" />
              </SelectTrigger>
              <SelectContent>
                {activeContacts.map((c) => (
                  <SelectItem key={c.Name} value={c.Name}>
                    {c.Name} ({c.Role})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="channel" className="text-xs">Channel</Label>
            <Select value={outreachChannel} onValueChange={setOutreachChannel}>
              <SelectTrigger id="channel">
                <SelectValue placeholder="Select channel" />
              </SelectTrigger>
              <SelectContent>
                {OUTREACH_CHANNELS.map((ch) => (
                  <SelectItem key={ch.value} value={ch.value}>
                    {ch.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="outreach-notes" className="text-xs">Notes</Label>
            <Textarea
              id="outreach-notes"
              placeholder="What happened..."
              value={outreachNotes}
              onChange={(e) => setOutreachNotes(e.target.value)}
              rows={2}
            />
          </div>
          <Button
            size="sm"
            className="w-full"
            disabled={loading || !outreachContact || !outreachChannel || !outreachNotes}
            onClick={handleOutreach}
          >
            {loading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : null}
            Log Outreach
          </Button>
        </CardContent>
      </Card>

      {/* Add Contact */}
      <Card className={!isActive ? "opacity-50 pointer-events-none" : ""}>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <UserPlus className="h-4 w-4" />
            Add Contact
          </CardTitle>
          <CardDescription>Add a new contact to reach out to</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="contact-name" className="text-xs">Name</Label>
              <Input
                id="contact-name"
                placeholder="Jane Smith"
                value={contactName}
                onChange={(e) => setContactName(e.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="contact-role" className="text-xs">Role</Label>
              <Input
                id="contact-role"
                placeholder="Staff Engineer"
                value={contactRole}
                onChange={(e) => setContactRole(e.target.value)}
              />
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="contact-linkedin" className="text-xs">LinkedIn URL</Label>
            <Input
              id="contact-linkedin"
              placeholder="https://linkedin.com/in/..."
              value={contactLinkedin}
              onChange={(e) => setContactLinkedin(e.target.value)}
            />
          </div>
          <Button
            size="sm"
            className="w-full"
            disabled={loading || !contactName || !contactRole}
            onClick={handleContact}
          >
            {loading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : null}
            Add Contact
          </Button>
        </CardContent>
      </Card>

      {/* Meeting Booked */}
      <Card className={!isActive ? "opacity-50 pointer-events-none" : ""}>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Trophy className="h-4 w-4" />
            Meeting Booked
          </CardTitle>
          <CardDescription>Mark this workflow as won</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-1.5">
            <Label htmlFor="meeting-notes" className="text-xs">Meeting Details</Label>
            <Textarea
              id="meeting-notes"
              placeholder="Meeting scheduled for..."
              value={meetingNotes}
              onChange={(e) => setMeetingNotes(e.target.value)}
              rows={2}
            />
          </div>
          <Button
            size="sm"
            variant="default"
            className="w-full bg-blue-600 hover:bg-blue-700 text-white"
            disabled={loading || !meetingNotes}
            onClick={handleMeetingBooked}
          >
            {loading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : null}
            Mark Meeting Booked
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
