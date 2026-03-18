import { createClient } from "@/lib/supabase/server"
import { redirect } from "next/navigation"

export default async function IntegrationsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 p-6">
      <h1 className="text-2xl font-semibold">Integrations</h1>
      <p className="text-muted-foreground">Signed in as {user.email}</p>
    </div>
  )
}
