const MAX_INPUT_LENGTH = 500

export async function generateTitle(
  userMessage: string,
  assistantReply: string
): Promise<string> {
  const claudeToken = process.env.CLAUDE_CODE_OAUTH_TOKEN
  if (!claudeToken) {
    return "Untitled"
  }

  const truncatedUser = userMessage.slice(0, MAX_INPUT_LENGTH)
  const truncatedReply = assistantReply.slice(0, MAX_INPUT_LENGTH)

  try {
    const res = await fetch("https://api.anthropic.com/v1/messages", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${claudeToken}`,
        "anthropic-version": "2023-06-01",
      },
      body: JSON.stringify({
        model: "claude-haiku-4-5",
        max_tokens: 32,
        messages: [
          {
            role: "user",
            content: `Generate a 4-6 word title for this conversation. Reply with ONLY the title, no quotes.\n\nUser: ${truncatedUser}\n\nAssistant: ${truncatedReply}`,
          },
        ],
      }),
    })

    if (!res.ok) {
      return "Untitled"
    }

    const json = await res.json()
    const text = json?.content?.[0]?.text?.trim()
    return text || "Untitled"
  } catch {
    return "Untitled"
  }
}
