"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function InstagramForm() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [fields, setFields] = useState({
    sessionId: "",
    csrfToken: "",
    dsUserId: "",
    mid: "",
    igDid: "",
    username: "",
  });

  function handleChange(key: keyof typeof fields, value: string) {
    setFields((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSave() {
    setSaving(true);
    const res = await fetch("/api/integrations/instagram/save", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(fields),
    });
    setSaving(false);
    if (res.ok) {
      setOpen(false);
      router.refresh();
    }
  }

  if (!open) {
    return (
      <Button size="sm" onClick={() => setOpen(true)}>
        Add Account
      </Button>
    );
  }

  return (
    <div className="space-y-3 rounded-lg border p-3">
      <p className="text-xs text-muted-foreground">
        Paste your Instagram cookies to connect.
      </p>
      {(
        [
          ["sessionId", "Session ID (sessionid)", true],
          ["csrfToken", "CSRF Token (csrftoken)", true],
          ["dsUserId", "DS User ID (ds_user_id)", true],
          ["mid", "MID (mid)", false],
          ["igDid", "IG DID (ig_did)", false],
          ["username", "Username (optional)", false],
        ] as [keyof typeof fields, string, boolean][]
      ).map(([key, label, required]) => (
        <div key={key} className="space-y-1">
          <Label htmlFor={`ig-${key}`} className="text-xs">
            {label}
            {required && <span className="text-destructive ml-0.5">*</span>}
          </Label>
          <Input
            id={`ig-${key}`}
            className="h-7 text-xs font-mono"
            value={fields[key]}
            onChange={(e) => handleChange(key, e.target.value)}
          />
        </div>
      ))}
      <div className="flex gap-2 pt-1">
        <Button size="sm" variant="outline" onClick={() => setOpen(false)}>
          Cancel
        </Button>
        <Button size="sm" onClick={handleSave} disabled={saving}>
          {saving ? "Saving…" : "Save"}
        </Button>
      </div>
    </div>
  );
}
