import { cookies } from "next/headers"

export async function getSessionCode(): Promise<string | null> {
  const cookieStore = await cookies()
  return cookieStore.get("engagent_session")?.value ?? null
}
