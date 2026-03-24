#!/usr/bin/env node

import { readFileSync } from "fs";
import { query } from "@anthropic-ai/claude-agent-sdk";

const sessionPath = process.argv[2];
if (!sessionPath) {
  console.error("Usage: node entrypoint.mjs <session.json>");
  process.exit(1);
}

const session = JSON.parse(readFileSync(sessionPath, "utf-8"));
const systemPrompt = session.systemPrompt || "You are a helpful assistant.";

const DIM = "\x1b[2m";
const RESET = "\x1b[0m";
const BOLD = "\x1b[1m";
const CYAN = "\x1b[36m";
const GREEN = "\x1b[32m";
const YELLOW = "\x1b[33m";
const RED = "\x1b[31m";
const GRAY = "\x1b[90m";

const write = (s) => process.stdout.write(s);

const conversation = query({
  prompt: session.prompt,
  options: {
    cwd: process.cwd(),
    permissionMode: "bypassPermissions",
    allowDangerouslySkipPermissions: true,
    systemPrompt,
    maxTurns: 500,
    model: session.model || "claude-sonnet-4-6",
  },
});

let turnCount = 0;

for await (const event of conversation) {
  if (event.type === "assistant" && event.message?.content) {
    for (const block of event.message.content) {
      if (block.type === "text") {
        // Agent's thinking/text — bold white
        write(`\n${BOLD}${block.text}${RESET}\n`);
      } else if (block.type === "tool_use") {
        turnCount++;
        const name = block.name || "tool";
        const input = block.input || {};

        if (name === "Bash") {
          // Show bash commands cleanly
          const cmd = (input.command || "").split("\n")[0].slice(0, 120);
          const desc = input.description || "";
          write(`\n${CYAN}[${turnCount}] ${desc}${RESET}\n`);
          write(`${DIM}  $ ${cmd}${cmd.length >= 120 ? "..." : ""}${RESET}\n`);
        } else if (name === "Write") {
          write(`\n${YELLOW}[${turnCount}] Writing ${input.file_path || "file"}${RESET}\n`);
        } else if (name === "Read") {
          write(`${GRAY}[${turnCount}] Reading ${input.file_path || "file"}${RESET}\n`);
        } else if (name === "Agent") {
          write(`\n${RED}[${turnCount}] Spawning agent: ${input.description || ""}${RESET}\n`);
        } else {
          const summary = JSON.stringify(input).slice(0, 100);
          write(`\n${CYAN}[${turnCount}] ${name}: ${summary}${RESET}\n`);
        }
      }
    }
  } else if (event.type === "tool_result") {
    // Show truncated result
    const raw = typeof event.content === "string"
      ? event.content
      : JSON.stringify(event.content || "");
    const output = raw.slice(0, 200);
    const lines = output.split("\n").slice(0, 3).join("\n    ");

    if (raw.toLowerCase().includes("error")) {
      write(`  ${RED}✗ ${lines}${RESET}\n`);
    } else {
      write(`  ${GREEN}✓${RESET} ${DIM}${lines}${raw.length > 200 ? "..." : ""}${RESET}\n`);
    }
  } else if (event.type === "result") {
    write(`\n${GREEN}${BOLD}━━━ Agent finished (${turnCount} tool calls) ━━━${RESET}\n`);
  }
}
