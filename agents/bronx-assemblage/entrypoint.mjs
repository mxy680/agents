#!/usr/bin/env node

import { readFileSync } from "fs";
import { query } from "@anthropic-ai/claude-agent-sdk";

const sessionPath = process.argv[2];
if (!sessionPath) {
  process.stderr.write("Usage: node entrypoint.mjs <session.json>\n");
  process.exit(1);
}

const session = JSON.parse(readFileSync(sessionPath, "utf-8"));
const systemPrompt = session.systemPrompt || "You are a helpful assistant.";

const conversation = query({
  prompt: session.prompt,
  options: {
    cwd: process.cwd(),
    permissionMode: "bypassPermissions",
    allowDangerouslySkipPermissions: true,
    systemPrompt,
    maxTurns: 30,
    model: session.model || "claude-sonnet-4-6",
  },
});

for await (const event of conversation) {
  if (event.type === "assistant" && event.message?.content) {
    for (const block of event.message.content) {
      if (block.type === "text") {
        process.stderr.write(block.text);
      }
    }
  } else if (event.type === "result") {
    process.stderr.write("\n---\nAgent finished.\n");
  }
}
