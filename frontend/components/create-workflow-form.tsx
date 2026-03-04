"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { Plus, Loader2 } from "lucide-react";

interface CreateWorkflowFormProps {
  onCreated: () => void;
}

export function CreateWorkflowForm({ onCreated }: CreateWorkflowFormProps) {
  const [companyName, setCompanyName] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!companyName.trim()) return;

    setLoading(true);
    try {
      const res = await fetch("/api/admin/companies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ companyName: companyName.trim() }),
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Failed to create workflow");
      }

      toast.success(`Workflow started for ${companyName.trim()}`);
      setCompanyName("");
      onCreated();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create workflow");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-3">
      <Input
        placeholder="Company name..."
        value={companyName}
        onChange={(e) => setCompanyName(e.target.value)}
        className="max-w-xs"
        disabled={loading}
      />
      <Button type="submit" disabled={loading || !companyName.trim()}>
        {loading ? (
          <Loader2 className="mr-1.5 h-4 w-4 animate-spin" />
        ) : (
          <Plus className="mr-1.5 h-4 w-4" />
        )}
        Start Workflow
      </Button>
    </form>
  );
}
