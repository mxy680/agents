"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

const COOKIE_FIELDS = [
  { name: "sessionId", label: "Session ID", required: true },
  { name: "csrfToken", label: "CSRF Token", required: true },
  { name: "dsUserId", label: "DS User ID", required: true },
  { name: "mid", label: "MID", required: false },
  { name: "igDid", label: "IG DID", required: false },
] as const;

export function InstagramForm() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [values, setValues] = useState({
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
      setValues({ sessionId: "", csrfToken: "", dsUserId: "", mid: "", igDid: "" });
      router.refresh();
    } else {
      const data = await res.json();
      setError(data.error ?? "Failed to save credentials");
    }
    setSaving(false);
  }

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
      >
        Enter Cookies
      </button>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      {COOKIE_FIELDS.map((field) => (
        <div key={field.name}>
          <label className="block text-xs font-medium text-gray-600 mb-1">
            {field.label}
            {field.required && <span className="text-red-500 ml-0.5">*</span>}
          </label>
          <input
            type="text"
            required={field.required}
            value={values[field.name]}
            onChange={(e) =>
              setValues({ ...values, [field.name]: e.target.value })
            }
            className="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-900 placeholder:text-gray-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
            placeholder={field.label}
          />
        </div>
      ))}
      {error && <p className="text-sm text-red-600">{error}</p>}
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={saving}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50 transition-colors"
        >
          {saving ? "Saving..." : "Save"}
        </button>
        <button
          type="button"
          onClick={() => setOpen(false)}
          className="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-600 hover:bg-gray-50 transition-colors"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
