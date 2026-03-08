"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useSignal } from "@/hooks/use-signal";
import { OUTREACH_CHANNELS } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { Mail, UserPlus, Trophy, CalendarIcon, Loader2 } from "lucide-react";
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

  const [meetingDate, setMeetingDate] = useState<Date>();
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
    await send("booked", { date: (meetingDate ?? new Date()).toISOString(), notes: meetingNotes });
    setMeetingDate(undefined);
    setMeetingNotes("");
  };

  const activeContacts = contacts.filter((c) => c.Active);

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">Signals</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="outreach">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="outreach" className="gap-1.5 text-xs">
              <Mail className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">Outreach</span>
            </TabsTrigger>
            <TabsTrigger value="contact" className="gap-1.5 text-xs">
              <UserPlus className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">Contact</span>
            </TabsTrigger>
            <TabsTrigger value="booked" className="gap-1.5 text-xs">
              <Trophy className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">Booked</span>
            </TabsTrigger>
          </TabsList>

          {/* Log Outreach */}
          <TabsContent value="outreach" className={!isActive ? "opacity-50 pointer-events-none" : ""}>
            <div className="space-y-3 pt-2">
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
            </div>
          </TabsContent>

          {/* Add Contact */}
          <TabsContent value="contact" className={!isActive ? "opacity-50 pointer-events-none" : ""}>
            <div className="space-y-3 pt-2">
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
            </div>
          </TabsContent>

          {/* Meeting Booked */}
          <TabsContent value="booked" className={!isActive ? "opacity-50 pointer-events-none" : ""}>
            <div className="space-y-3 pt-2">
              <div className="space-y-1.5">
                <Label className="text-xs">Meeting Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal",
                        !meetingDate && "text-muted-foreground"
                      )}
                    >
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {meetingDate ? format(meetingDate, "PPP") : "Pick a date"}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={meetingDate}
                      onSelect={setMeetingDate}
                    />
                  </PopoverContent>
                </Popover>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="meeting-notes" className="text-xs">Notes (optional)</Label>
                <Textarea
                  id="meeting-notes"
                  placeholder="Any details about the meeting..."
                  value={meetingNotes}
                  onChange={(e) => setMeetingNotes(e.target.value)}
                  rows={2}
                />
              </div>
              <Button
                size="sm"
                variant="default"
                className="w-full bg-blue-600 hover:bg-blue-700 text-white"
                disabled={loading || !meetingDate}
                onClick={handleMeetingBooked}
              >
                {loading ? <Loader2 className="mr-1.5 h-4 w-4 animate-spin" /> : null}
                Mark Meeting Booked
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
