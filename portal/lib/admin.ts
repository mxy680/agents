const ADMIN_EMAILS = (process.env.ADMIN_EMAILS ?? "")
  .split(",")
  .map((e) => e.trim())
  .filter(Boolean)

export function isAdmin(email: string | undefined): boolean {
  return !!email && ADMIN_EMAILS.includes(email)
}
