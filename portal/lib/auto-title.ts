const MAX_TITLE_LENGTH = 50

/**
 * Generate a conversation title from the first user message.
 * Truncates to ~50 chars at a word boundary.
 */
export function generateTitle(userMessage: string): string {
  const cleaned = userMessage.trim().replace(/\n+/g, " ")
  if (cleaned.length <= MAX_TITLE_LENGTH) return cleaned
  const truncated = cleaned.slice(0, MAX_TITLE_LENGTH)
  const lastSpace = truncated.lastIndexOf(" ")
  return (lastSpace > 20 ? truncated.slice(0, lastSpace) : truncated) + "…"
}
