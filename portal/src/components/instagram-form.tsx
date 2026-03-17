"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const COOKIE_FIELDS = [
  { name: "sessionId", label: "Session ID", required: true },
  { name: "csrfToken", label: "CSRF Token", required: true },
  { name: "dsUserId", label: "DS User ID", required: true },
  { name: "mid", label: "MID", required: false },
  { name: "igDid", label: "IG DID", required: false },
] as const;

type CookieFieldName = (typeof COOKIE_FIELDS)[number]["name"];

export function InstagramForm() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [values, setValues] = useState<Record<CookieFieldName | "username", string>>({
    username: "",
    sessionId: "",
    csrfToken: "",
    dsUserId: "",
    mid: "",
    igDid: "",
  });

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);

    const res = await fetch("/api/integrations/instagram/save", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(values),
    });

    if (res.ok) {
      setOpen(false);
      setValues({ username: "", sessionId: "", csrfToken: "", dsUserId: "", mid: "", igDid: "" });
      router.refresh();
    } else {
      const data = await res.json();
      setError(data.error ?? "Failed to save credentials");
    }
    setSaving(false);
  }

  if (!open) {
    return (
      <Button size="sm" onClick={() => setOpen(true)}>
        Enter Cookies
      </Button>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div className="space-y-1">
        <Label htmlFor="username" className="text-xs">
          Username (account label, optional)
        </Label>
        <Input
          id="username"
          type="text"
          value={values.username}
          onChange={(e) => setValues({ ...values, username: e.target.value })}
          placeholder="your_instagram_username"
          className="h-8 text-sm"
        />
      </div>
      {COOKIE_FIELDS.map((field) => (
        <div key={field.name} className="space-y-1">
          <Label htmlFor={field.name} className="text-xs">
            {field.label}
            {field.required && <span className="text-destructive ml-0.5">*</span>}
          </Label>
          <Input
            id={field.name}
            type="text"
            required={field.required}
            value={values[field.name]}
            onChange={(e) =>
              setValues({ ...values, [field.name]: e.target.value })
            }
            placeholder={field.label}
            className="h-8 text-sm"
          />
        </div>
      ))}
      {error && <p className="text-sm text-destructive">{error}</p>}
      <div className="flex gap-2">
        <Button type="submit" size="sm" disabled={saving}>
          {saving ? "Saving..." : "Save"}
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setOpen(false)}
        >
          Cancel
        </Button>
      </div>
    </form>
  );
}
