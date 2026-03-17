"use client";

import { buttonVariants } from "@/components/ui/button";

export function ConnectButton({ href }: { href: string }) {
  return (
    <a href={href} className={buttonVariants({ size: "sm" })}>
      Add Account
    </a>
  );
}
