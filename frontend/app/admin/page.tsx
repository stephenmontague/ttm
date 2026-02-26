"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { listCompanies, createCompany } from "@/lib/api";
import { Company } from "@/lib/types";

export default function AdminPage() {
  const [companies, setCompanies] = useState<Company[]>([]);
  const [newCompany, setNewCompany] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    loadCompanies();
  }, []);

  async function loadCompanies() {
    try {
      const data = await listCompanies();
      setCompanies(data);
    } catch (err) {
      setError("Failed to load companies");
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!newCompany.trim()) return;

    setLoading(true);
    setError("");
    try {
      await createCompany(newCompany.trim());
      setNewCompany("");
      // Wait a moment for the snapshot to be written, then refresh
      setTimeout(loadCompanies, 1000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create workflow");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="container mx-auto px-4 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Admin Dashboard</h1>
            <p className="text-sm text-muted-foreground">
              Manage outreach workflows
            </p>
          </div>
          <Link
            href="/"
            className="text-sm text-muted-foreground hover:text-foreground"
          >
            Public View
          </Link>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8 space-y-8">
        {error && (
          <div className="bg-destructive/10 text-destructive px-4 py-2 rounded-md text-sm">
            {error}
          </div>
        )}

        <Card>
          <CardHeader>
            <CardTitle>Start New Workflow</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreate} className="flex gap-2">
              <Input
                placeholder="Company name (e.g. Whoop)"
                value={newCompany}
                onChange={(e) => setNewCompany(e.target.value)}
              />
              <Button type="submit" disabled={loading}>
                {loading ? "Starting..." : "Start Workflow"}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>All Workflows</CardTitle>
          </CardHeader>
          <CardContent>
            {companies.length === 0 ? (
              <p className="text-muted-foreground text-sm">
                No workflows yet. Create one above.
              </p>
            ) : (
              <div className="space-y-2">
                {companies.map((c) => (
                  <Link
                    key={c.ID}
                    href={`/admin/company/${c.Slug}`}
                    className="flex items-center justify-between p-3 rounded-md border hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <span className="font-medium">{c.CompanyName}</span>
                      <Badge variant="outline">{c.Status}</Badge>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {c.OutreachCount} attempts &middot; {c.ElapsedDays}d
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
