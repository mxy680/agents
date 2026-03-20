const ADMIN_EMAILS = (process.env.ADMIN_EMAILS ?? "markshteyn1@gmail.com")
  .split(",")
  .map((e) => e.trim())

export function isAdmin(email: string | undefined): boolean {
  return !!email && ADMIN_EMAILS.includes(email)
}
