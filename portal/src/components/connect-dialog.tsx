"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

export function ConnectDialog({
  href,
  providerName,
}: {
  href: string;
  providerName: string;
}) {
  const [open, setOpen] = useState(false);
  const [accountName, setAccountName] = useState("");

  function handleConnect() {
    const url = accountName.trim()
      ? `${href}?label=${encodeURIComponent(accountName.trim())}`
      : href;
    window.location.href = url;
  }

  return (
    <>
      <Button size="sm" onClick={() => setOpen(true)}>
        Add Account
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Connect {providerName}</DialogTitle>
            <DialogDescription>
              Give this account a name so you can identify it later.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2 py-2">
            <Label htmlFor="account-name">Account name</Label>
            <Input
              id="account-name"
              placeholder="e.g. Work, Personal"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleConnect()}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button size="sm" onClick={handleConnect}>
              Continue
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
