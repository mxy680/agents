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
    maxTurns: 200,
    model: session.model || "claude-sonnet-4-6",
  },
});

for await (const event of conversation) {
  if (event.type === "assistant" && event.message?.content) {
    for (const block of event.message.content) {
      if (block.type === "text") {
        process.stderr.write(block.text);
      } else if (block.type === "tool_use") {
        const name = block.name || "tool";
        const input = typeof block.input === "string"
          ? block.input.slice(0, 200)
          : JSON.stringify(block.input || {}).slice(0, 200);
        process.stderr.write(`\n[${name}] ${input}\n`);
      }
    }
  } else if (event.type === "tool_result") {
    const output = typeof event.content === "string"
      ? event.content
      : JSON.stringify(event.content || "").slice(0, 300);
    process.stderr.write(`  → ${output.slice(0, 300)}\n`);
  } else if (event.type === "result") {
    process.stderr.write("\n---\nAgent finished.\n");
  }
}
