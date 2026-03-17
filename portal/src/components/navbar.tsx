import Link from "next/link";
import { createClient } from "@/lib/supabase/server";
import { SignOutButton } from "./sign-out-button";

export async function Navbar() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  return (
    <nav className="border-b border-gray-200 bg-white">
      <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-4">
        <Link href="/" className="text-lg font-bold text-gray-900">
          Agent Marketplace
        </Link>
        <div className="flex items-center gap-4">
          {user ? (
            <>
              <Link
                href="/integrations"
                className="text-sm text-gray-600 hover:text-gray-900"
              >
                Integrations
              </Link>
              <span className="text-sm text-gray-500">{user.email}</span>
              <SignOutButton />
            </>
          ) : (
            <Link
              href="/login"
              className="rounded-md bg-gray-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
            >
              Sign in
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
